package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/js13kgames/glitchd/server"
	"github.com/js13kgames/glitchd/server/log"
)

const motd = `

            dP oo   dP            dP             dP
            88      88            88             88
   .d8888b. 88 dP d8888P .d8888b. 88d888b. .d888b88      * glitchd %s
   88'  '88 88 88   88   88'  '"" 88'  '88 88'  '88      * Built using %s, maxprocs = %d
   88.  .88 88 88   88   88.  ... 88    88 88.  .88
   '8888P88 dP dP   dP   '88888P' dP    dP '88888P8
   .88
   d8888P

`

func main() {
	// @todo We don't have proper configs currently, so loglevels are build flag dependent and hardcoded.
	logger, err := log.New()
	if err != nil {
		os.Stderr.Write([]byte(err.Error()))
		os.Exit(1)
	}
	log.SetLevel(loglevel)

	// Print our "pretty" logo.
	fmt.Fprintf(os.Stdout, motd, server.Version, runtime.Version(), runtime.GOMAXPROCS(0))

	(&Runner{
		closing: make(chan struct{}),
		Closed:  make(chan struct{}),
		logger:  logger,
	}).Run()
}
