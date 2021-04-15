package process

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// PollResult can be process or error
type PollResult struct {
	Process *Process
	Error   error
}

var nsPidRegExp *regexp.Regexp
var vmPeakExp *regexp.Regexp
var vmSizeExp *regexp.Regexp
var vmLckExp *regexp.Regexp
var vmPinExp *regexp.Regexp
var vmHWMExp *regexp.Regexp
var vmRSSExp *regexp.Regexp
var rssAnonExp *regexp.Regexp
var rssFileExp *regexp.Regexp
var rssShmemExp *regexp.Regexp

// Poll polls processes with specified interval and writes them to channel
func Poll(c chan PollResult, interval time.Duration) {
	var err error

	rand.Seed(time.Now().UnixNano())

	nsPidRegExp, err = regexp.Compile("^NSpid:.*\\s(\\d+)\\s*$")
	if err != nil {
		c <- PollResult{Error: err}
		return
	}

	vmPeakExp, err = regexp.Compile("^VmPeak:.*\\s(\\d+)\\s*kB")
	if err != nil {
		c <- PollResult{Error: err}
		return
	}

	vmSizeExp, err = regexp.Compile("^VmSize:.*\\s(\\d+)\\s*kB")
	if err != nil {
		c <- PollResult{Error: err}
		return
	}

	vmLckExp, err = regexp.Compile("^VmLck:.*\\s(\\d+)\\s*kB")
	if err != nil {
		c <- PollResult{Error: err}
		return
	}

	vmPinExp, err = regexp.Compile("^VmPin:.*\\s(\\d+)\\s*kB")
	if err != nil {
		c <- PollResult{Error: err}
		return
	}

	vmHWMExp, err = regexp.Compile("^VmHWM:.*\\s(\\d+)\\s*kB")
	if err != nil {
		c <- PollResult{Error: err}
		return
	}

	vmRSSExp, err = regexp.Compile("^VmRSS:.*\\s(\\d+)\\s*kB")
	if err != nil {
		c <- PollResult{Error: err}
		return
	}

	rssAnonExp, err = regexp.Compile("^RssAnon:.*\\s(\\d+)\\s*kB")
	if err != nil {
		c <- PollResult{Error: err}
		return
	}

	rssFileExp, err = regexp.Compile("^RssFile:.*\\s(\\d+)\\s*kB")
	if err != nil {
		c <- PollResult{Error: err}
		return
	}

	rssShmemExp, err = regexp.Compile("^RssShmem:.*\\s(\\d+)\\s*kB")
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
	seqID := rand.Int31()

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
				p.SeqID = seqID
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

	err = fillProcessCgroup(pid, p)
	if err != nil {
		return nil, err
	}

	err = fillProcessStatus(pid, p)
	if err != nil {
		log.Print(err)
	}

	err = fillProcessCmd(pid, p)
	if err != nil {
		log.Print(err)
	}

	return p, nil
}

func fillProcessCgroup(pid uint64, p *Process) error {
	f, err := os.Open(fmt.Sprint("/proc/", pid, "/cgroup"))
	if err != nil {
		return err
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
			return err
		}

		if len(record) < 3 {
			continue
		}

		cgroupCtrl, cgroup := record[1], record[2]
		if cgroupCtrl == "pids" {
			p.Cgroup = cgroup
			return nil
		}
	}

	return fmt.Errorf("no cgroup for %d", pid)
}

func fillProcessStatus(pid uint64, p *Process) error {
	f, err := os.Open(fmt.Sprint("/proc/", pid, "/status"))
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		var match []string

		match = nsPidRegExp.FindStringSubmatch(line)
		if match != nil {
			p.NSpid, _ = strconv.ParseUint(match[1], 10, 64)
		}

		match = vmPeakExp.FindStringSubmatch(line)
		if match != nil {
			p.VmPeak, _ = strconv.ParseUint(match[1], 10, 64)
		}

		match = vmSizeExp.FindStringSubmatch(line)
		if match != nil {
			p.VmSize, _ = strconv.ParseUint(match[1], 10, 64)
		}

		match = vmLckExp.FindStringSubmatch(line)
		if match != nil {
			p.VmLck, _ = strconv.ParseUint(match[1], 10, 64)
		}

		match = vmPinExp.FindStringSubmatch(line)
		if match != nil {
			p.VmPin, _ = strconv.ParseUint(match[1], 10, 64)
		}

		match = vmHWMExp.FindStringSubmatch(line)
		if match != nil {
			p.VmHWM, _ = strconv.ParseUint(match[1], 10, 64)
		}

		match = vmRSSExp.FindStringSubmatch(line)
		if match != nil {
			p.VmRSS, _ = strconv.ParseUint(match[1], 10, 64)
		}

		match = rssAnonExp.FindStringSubmatch(line)
		if match != nil {
			p.RssAnon, _ = strconv.ParseUint(match[1], 10, 64)
		}

		match = rssFileExp.FindStringSubmatch(line)
		if match != nil {
			p.RssFile, _ = strconv.ParseUint(match[1], 10, 64)
		}

		match = rssShmemExp.FindStringSubmatch(line)
		if match != nil {
			p.RssShmem, _ = strconv.ParseUint(match[1], 10, 64)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func fillProcessCmd(pid uint64, p *Process) error {
	data, err := ioutil.ReadFile(fmt.Sprint("/proc/", pid, "/cmdline"))
	if err != nil {
		return err
	}

	p.Cmd = strings.ReplaceAll(string(data), "\x00", "")
	return nil
}
