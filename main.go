package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"

	"github.com/gophergala/nedomi/app"
	"github.com/gophergala/nedomi/config"
	"github.com/gophergala/nedomi/utils"
)

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

	cfg, err := config.Get()

	if err != nil {
		log.Fatalf("Error parsing config. %s\n", err)
	}

	if testConfig {
		os.Exit(0)
	}

	if debug {
		cfg.Logging.Debug = debug
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

	if !debug && cfg.Logging.LogFile != "" {
		logFile, err := os.OpenFile(cfg.Logging.LogFile,
			os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0665)
		if err != nil {
			log.Fatalf("Error setting logfile. %s", err)
		}
		log.SetOutput(logFile)
		defer logFile.Close()
	}

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
