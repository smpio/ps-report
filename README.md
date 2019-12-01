# ps-report

This daemon listens for process fork events using kernel netlink interface and sends process cgroup information to redis. This will help you to debug OOM events in a cluster of linux containers. OOM event only contains PID and process executable name without container name.

## Requiremenets

* Kernel should be build with `CONFIG_PROC_EVENTS=y` (check in `/proc/config.gz`)
* The program should be run in privileged mode with host network and host pid namespace (`docker run --privileged=true --net=host --pid=host`)

## Usage

```
TODO: paste ./ps-report -h
```

## See

* https://github.com/kinvolk/nswatch
* https://github.com/cloudfoundry/gosigar/tree/master/psnotify
* https://godoc.org/github.com/remyoudompheng/go-netlink
