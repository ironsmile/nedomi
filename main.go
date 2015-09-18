// nedomi is a HTTP media caching server. It aims to increase performance by
// choosing chaching algorithms suitable for media files.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/ironsmile/nedomi/app"
	"github.com/ironsmile/nedomi/config"
)

const (
	// Version will be reported if the -v flag is used
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

//!TODO: implement some "unit" tests for this :)
func run() int {
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not create cpuprofile file: %s\n", err)
			return 1
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			panic(err)
		}
		defer pprof.StopCPUProfile()
	}

	if showVersion {
		fmt.Printf("nedomi version %s\n", Version)
		return 0
	}

	cfg, err := config.Get()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing config: %s\n", err)
		return 2
	}

	if testConfig {
		return 0
	}

	//!TODO: simplify and encapsulate application startup:
	// Move/encapsulate SetupEnv/CleanupEnv, New, Start, Wait, etc.
	// Leave only something like return App.Run(cfg)
	// This will possibly simplify configuration reloading and higher contexts as well
	defer func() {
		if err := app.CleanupEnv(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't cleunup after nedomi: %s\n", err)
		}
	}()
	if err := app.SetupEnv(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't setup nedomi environment: %s\n", err)
		return 3
	}

	appInstance, err := app.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't initialize nedomi: %s\n", err)
		return 4
	}

	if err := appInstance.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't start nedomi: %s\n", err)
		return 5
	}

	if err := appInstance.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "Error stopping the app : %s\n", err)
		return 6
	}

	return 0
}

func main() {
	flag.Parse()

	os.Exit(run())
}
