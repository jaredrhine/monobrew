package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/jaredrhine/monobrew/pkg/monobrew"
)

type configList []string

var verbose bool
var veryVerbose bool
var nuke bool
var configFiles configList

func init() {
	const (
		defaultVerbose     = false
		defaultVeryVerbose = false
		defaultNuke        = false
	)

	flag.BoolVar(&verbose, "v", defaultVerbose, "verbose printing")
	flag.BoolVar(&veryVerbose, "vv", defaultVeryVerbose, "debug printing")
	flag.BoolVar(&nuke, "nuke", defaultNuke, "nuke (completely remove all contents) the state-dir directory before start")
	flag.Var(&configFiles, "config", "add a config file")
}

func (cl *configList) String() string {
	return strings.Join(*cl, " ")
}

func (cl *configList) Set(value string) error {
	*cl = append(*cl, strings.TrimSpace(value))
	return nil
}

func main() {
	flag.Parse()

	if len(configFiles) == 0 {
		fmt.Println("you must specify a `--config /path/to/configfile` argument")
		return
	}

	config := monobrew.NewConfig()
	config.PrintVerboseResult = verbose
	config.PrintDebug = veryVerbose
	config.NukeStateDirAtStart = nuke
	config.ConfigFiles = configFiles
	config.Load()

	runner := monobrew.NewRunner(config)
	runner.RunOps()
}
