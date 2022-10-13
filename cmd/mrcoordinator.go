package main

import (
	"fmt"
	"github/rolandvarga/go-mapreduce/internal/coordinator"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: mrcoordinator inputfiles...\n")
		os.Exit(1)
	}

	// keeping logger setup here, since main file can't be modified
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	m := coordinator.MakeCoordinator(os.Args[1:], 10)
	for m.Done() == false {
		time.Sleep(time.Second)
	}

	time.Sleep(time.Second)
}
