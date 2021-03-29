package process

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"time"
)

// PollResult can be process or error
type PollResult struct {
	Process *Process
	Error   error
}

var nsPIDRegExp *regexp.Regexp

// Poll polls processes with specified interval and writes them to channel
func Poll(c chan PollResult, interval time.Duration) {
	var err error
	nsPIDRegExp, err = regexp.Compile("^NSpid:.*\\s(\\d+)\\s*$")
	if err != nil {
		c <- PollResult{Error: err}
		return
	}

	for {
		getProcesses(c)
		time.Sleep(interval)
	}
}

func getProcesses(c chan PollResult) {
	d, err := os.Open("/proc")
	if err != nil {
		c <- PollResult{Error: err}
		return
	}
	defer d.Close()

	for {
		fis, err := d.Readdir(10)
		if err == io.EOF {
			break
		}
		if err != nil {
			c <- PollResult{Error: err}
			return
		}

		for _, fi := range fis {
			// We only care about directories, since all pids are dirs
			if !fi.IsDir() {
				continue
			}

			name := fi.Name()
			pid, err := strconv.ParseUint(name, 10, 64)
			if err != nil {
				continue
			}

			p, err := makeProcess(pid)
			if err == nil {
				c <- PollResult{Process: p}
			} else {
				c <- PollResult{Error: err}
			}
		}
	}
}

func makeProcess(pid uint64) (*Process, error) {
	var err error
	p := &Process{Pid: pid}
	p.Cgroup, err = getProcessCgroup(pid)
	if err == nil {
		p.NSPID, _ = getProcessContainerPID(pid)
	}

	return p, err
}

func getProcessCgroup(pid uint64) (string, error) {
	f, err := os.Open(fmt.Sprint("/proc/", pid, "/cgroup"))
	if err != nil {
		return "", err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = ':'

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		if len(record) < 3 {
			continue
		}

		cgroupCtrl, cgroup := record[1], record[2]
		if cgroupCtrl == "pids" {
			return cgroup, nil
		}
	}

	return "", fmt.Errorf("no cgroup for %d", pid)
}

func getProcessContainerPID(pid uint64) (uint64, error) {
	f, err := os.Open(fmt.Sprint("/proc/", pid, "/status"))
	if err != nil {
		return 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		match := nsPIDRegExp.FindStringSubmatch(line)
		if match != nil {
			nsPID, err := strconv.ParseUint(match[1], 10, 64)
			if err != nil {
				return 0, err
			}
			return nsPID, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return 0, fmt.Errorf("no NSpid for %d", pid)
}
