package main

import (
	"flag"
	"fmt"

	"github.com/jaredrhine/monobrew/pkg/monobrew"
)

var verbose bool
var veryVerbose bool
var nuke bool
var configFile string

func init() {
	const (
		defaultVerbose     = false
		defaultVeryVerbose = false
		defaultNuke        = false
	)

	flag.BoolVar(&verbose, "v", defaultVerbose, "verbose printing")
	flag.BoolVar(&veryVerbose, "vv", defaultVeryVerbose, "debug printing")
	flag.BoolVar(&nuke, "nuke", defaultNuke, "nuke (completely remove all contents) the state-dir directory before start")
	flag.StringVar(&configFile, "config", "", "add a config file")
}

func main() {
	flag.Parse()

	if configFile == "" {
		fmt.Println("you must specify a `--config /path/to/configfile` argument")
		return
	}

	config := monobrew.NewConfig()
	config.PrintVerboseResult = verbose
	config.PrintDebug = veryVerbose
	config.NukeStateDirAtStart = nuke
	config.AddConfigFile(string(configFile))
	config.Load()

	runner := monobrew.NewRunner(config)
	runner.RunOps()
}
