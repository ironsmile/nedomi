// nedomi is a HTTP media caching server. It aims to increase performance by
// choosing chaching algorithms suitable for media files.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/ironsmile/nedomi/app"
	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
)

var (
	// Version will be reported if the -v flag is used
	Version = "not build through make"
	// BuildTime is the build time of the binary
	BuildTime string
	// GitHash is the commit hash from which the version is build
	GitHash string
	// GitTag the tag (if any) of the commit
	// from which the version is build
	GitTag string
	// Dirty is flag telling whether th e build is dirty - build in part from uncommitted code
	// setting it to 'true' means that it is.
	Dirty string
)

// The following will be populated from the command line with via `flag`
var (
	testConfig  bool
	showVersion bool
	cpuprofile  string
)

func init() {
	flag.BoolVar(&testConfig, "t", false, "Test configuration file and exit")
	flag.BoolVar(&showVersion, "v", false, "Print version information")
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
	// error probably means this is not build with make
	n, _ := strconv.ParseInt(BuildTime, 10, 64)
	buildTime := time.Unix(n, 0)
	var appVersion = types.AppVersion{
		Version:   Version,
		GitHash:   GitHash,
		GitTag:    GitTag,
		BuildTime: buildTime,
		Dirty:     Dirty == "true",
	}

	if showVersion {
		fmt.Printf("nedomi version %s\n", appVersion)
		return 0
	}

	appInstance, err := app.New(appVersion, config.Get)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't initialize nedomi: %s\n", err)
		return 4
	}

	if testConfig {
		return 0 // still doesn't work :)
	}

	if err := appInstance.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Nedomi exit with error: %s\n", err)
		return 5
	}

	return 0
}

func absolutizeArgv0() error {
	if filepath.IsAbs(os.Args[0]) {
		return nil
	}
	argv0, err := exec.LookPath(os.Args[0])
	if err != nil {
		return err
	}
	argv0, err = filepath.Abs(argv0)
	if err != nil {
		return err
	}
	os.Args[0] = argv0
	return nil
}

func main() {
	if err := absolutizeArgv0(); err != nil {
		fmt.Fprintf(os.Stderr, "Error while absolutizing the path to the executable: %s\n", err)
		os.Exit(9)
	}

	flag.Parse()

	os.Exit(run())
}
