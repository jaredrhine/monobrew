package main

import (
	"flag"
	"fmt"

	"github.com/jaredrhine/monobrew/pkg/monobrew"
)

var defaultPrintVerboseResult bool

func init() {
	defaultPrintVerboseResult = false
}

func main() {
	configFilePtr := flag.String("config", "", "config file")
	flag.Parse()

	if *configFilePtr == "" {
		fmt.Println("you must specify a `--config /path/to/configfile` argument")
		return
	}

	config := monobrew.NewConfig()
	config.PrintVerboseResult = defaultPrintVerboseResult
	config.AddConfigFile(string(*configFilePtr))
	config.Init()

	fmt.Printf("%#v\n", config)
	runner := monobrew.NewRunner(config)
	runner.Scan()
	runner.RunOps()
}
