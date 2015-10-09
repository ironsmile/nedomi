// nedomi is a HTTP media caching server. It aims to increase performance by
// choosing chaching algorithms suitable for media files.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/pprof"

	"github.com/ironsmile/nedomi/app"
	"github.com/ironsmile/nedomi/config"
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

func buildVersion() string {
	var ver = bytes.NewBufferString(Version)
	if GitTag != "" {
		ver.WriteRune('-')
		ver.WriteString(GitTag)
	} else if GitHash != "" {
		ver.WriteRune('-')
		ver.WriteString(GitHash)
	}

	if BuildTime != "" {
		ver.WriteString(" build at ")
		ver.WriteString(BuildTime)
	}
	return ver.String()
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
		fmt.Printf("nedomi version %s\n", buildVersion())
		return 0
	}
	if err := absolutizeArgv0(); err != nil {
		fmt.Fprintf(os.Stderr, "Error while absolutizing the path to the executable: %s\n", err)
		return 9
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
	flag.Parse()

	os.Exit(run())
}
