/*
	nedomi is a HTTP Media cache. It aims to increase performance with
	choosing chaching algorithms suitable for media files.
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"

	"github.com/ironsmile/nedomi/app"
	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/utils"
)

const (
	// This will be reported as a version of the software
	Version = "alpha-1-development"
)

// The following will be populated from the command line with via `flag`
var (
	testConfig  bool
	showVersion bool
	debug       bool
	cpuprofile  string
)

func init() {
	flag.BoolVar(&testConfig, "t", false, "Test configuration file and exit")
	flag.BoolVar(&showVersion, "v", false, "Print version information")
	flag.BoolVar(&debug, "D", false, "Debug. Will print messages into stdout")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "Write cpu profile to this file")

	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	flag.Parse()

	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatalf("Creating cpuprofile file. %s", err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if showVersion {
		fmt.Printf("nedomi version %s\n", Version)
		os.Exit(0)
	}

	cfg, err := config.Get()

	if err != nil {
		log.Fatalf("Error parsing config. %s\n", err)
	}

	if testConfig {
		os.Exit(0)
	}

	if absPath, err := filepath.Abs(config.ConfigFile); err != nil {
		log.Fatalf("Was not able to find config absolute path. %s", err)
	} else {
		config.ConfigFile = absPath
	}

	err = utils.SetupEnv(cfg)

	if err != nil {
		log.Fatalln(err)
	}

	defer utils.CleanupEnv(cfg)

	appl, err := app.New(cfg)

	if err != nil {
		utils.CleanupEnv(cfg)
		log.Fatalln(err)
	}

	if err = appl.Start(); err != nil {
		utils.CleanupEnv(cfg)
		log.Fatalln(err)
	}

	if err := appl.Wait(); err != nil {
		log.Printf("Error stopping the app : %s\n", err)
		os.Exit(1)
	}
}
