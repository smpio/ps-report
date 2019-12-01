# ps-report

This daemon periodically collects information for all running processes and sends cgroup information to postgres. This will help you to debug OOM events in a cluster of linux containers. OOM event only contains PID and process executable name without container name.

## Requiremenets

* The program should be run with host pid namespace (`docker run --pid=host`)

## Usage

```
  -db-url string
    	database URL
  -hostname string
    	hostname used in redis key prefix (default "docker-desktop")
  -poll-interval uint
    	process poll interval in seconds (default 60)
```
