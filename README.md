# ps-report

This daemon periodically collects information for all running processes and sends cgroup information to redis. This will help you to debug OOM events in a cluster of linux containers. OOM event only contains PID and process executable name without container name.

## Requiremenets

* The program should be run with host pid namespace (`docker run --pid=host`)

## Usage

```
  -hostname string
    	hostname used in redis key prefix (default is hostname returned by os)
  -poll-interval uint
    	process poll interval in seconds (default 60)
  -redis-addr string
    	redis address in form host:port
  -redis-db int
    	redis DB
  -redis-passwd string
    	redis password
```
