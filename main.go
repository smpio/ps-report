package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/smpio/ps-report/datastore"
	"github.com/smpio/ps-report/process"
)

func main() {
	defaultHostname, err := os.Hostname()

	dbURL := flag.String("db-url", "", "database URL")
	pollIntervalSeconds := flag.Uint("poll-interval", 60, "process poll interval in seconds")
	hostname := flag.String("hostname", defaultHostname, "hostname used in redis key prefix")
	flag.Parse()

	if *dbURL == "" {
		log.Fatalln("Database URL not set")
	}

	ds, err := datastore.New(*dbURL)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Connected to database")

	pollInterval := time.Duration(*pollIntervalSeconds) * time.Second

	c := make(chan process.PollResult, 1024)
	go process.Poll(c, pollInterval)

	for res := range c {
		if res.Error != nil {
			log.Println(res.Error)
		}
		p := res.Process

		if err := ds.Write(*hostname, p.Pid, p.Cgroup); err != nil {
			log.Println(err)
		}
	}

	ds.Close()
}
