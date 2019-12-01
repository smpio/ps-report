package process

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

// PollResult can be process or error
type PollResult struct {
	Process *Process
	Error   error
}

// PollProcesses polls processes with specified interval and writes them to channel
func Poll(c chan PollResult, interval time.Duration) {
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

			// We only care if the name starts with a numeric
			name := fi.Name()
			if name[0] < '0' || name[0] > '9' {
				continue
			}

			pid, err := strconv.ParseInt(name, 10, 0)
			if err != nil {
				continue
			}

			p, err := makeProcess(pid)
			if err != nil {
				c <- PollResult{Error: err}
			}

			c <- PollResult{Process: p}
		}
	}
}

func makeProcess(pid int64) (*Process, error) {
	p := &Process{Pid: pid}
	cgroup, err := getProcessCgroup(pid)
	if err != nil {
		return p, err
	}

	p.Cgroup = cgroup
	return p, nil
}

func getProcessCgroup(pid int64) (string, error) {
	f, err := os.Open(fmt.Sprint("/proc/", pid, "/cgroup"))
	if err != nil {
		return "", err
	}

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