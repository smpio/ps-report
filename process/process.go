package process

// Process pid and cgroup
type Process struct {
	Pid    uint64
	Cgroup string
	NSPID  uint64
}
