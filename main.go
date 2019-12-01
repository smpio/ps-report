package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/smpio/ps-report/datastore"
	"github.com/smpio/ps-report/process"
)

func main() {
	defaultHostname, err := os.Hostname()

	redisAddr := flag.String("redis-addr", "", "redis address in form host:port")
	redisPasswd := flag.String("redis-passwd", "", "redis password")
	redisDB := flag.Int("redis-db", 0, "redis DB")
	pollIntervalSeconds := flag.Uint("poll-interval", 60, "process poll interval in seconds")
	hostname := flag.String("hostname", defaultHostname, "hostname used in redis key prefix")
	flag.Parse()

	if *redisAddr == "" {
		log.Fatalln("Redis address not set")
	}

	ds, err := datastore.New(*redisAddr, *redisPasswd, *redisDB)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Connected to Redis")

	pollInterval := time.Duration(*pollIntervalSeconds) * time.Second
	recordExpiration := pollInterval + (5 * time.Second)

	c := make(chan process.PollResult, 1024)
	go process.Poll(c, pollInterval)

	for res := range c {
		if res.Error != nil {
			log.Println(res.Error)
		}
		p := res.Process

		if err := ds.Write(fmt.Sprint(*hostname, ":", p.Pid, ":", p.Cgroup), "", recordExpiration); err != nil {
			log.Println(err)
		}
	}

	ds.Close()
}
