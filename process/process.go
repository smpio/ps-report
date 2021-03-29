package process

// Process pid and cgroup
type Process struct {
	SeqID       int32
	Pid         uint64
	Cgroup      string
	NSpid       uint64
	VmPeak      uint64
	VmSize      uint64
	VmLck       uint64
	VmPin       uint64
	VmHWM       uint64
	VmRSS       uint64
	RssAnon     uint64
	RssFile     uint64
	RssShmem    uint64
}
