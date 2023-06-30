package monobrew

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Runner struct {
	Config    *Config
	OpCounter uint
}

func NewRunner(config *Config) *Runner {
	return &Runner{Config: config}
}

func (r *Runner) RunOps() {
	var ops []*Block

	config := r.Config
	ops = config.OrderedOps()

	for i, op := range ops {
		r.OpCounter += 1
		op.OpCounter = r.OpCounter
		ExecuteOp(op)
		r.DumpState(op)

		if config.PrintVerboseResult {
			if i != 0 {
				fmt.Println()
			}
			fmt.Printf("--- op #%v --> %v ---------------\n", r.OpCounter, op.Label)
			fmt.Printf("command: %s %s\n", op.CommandPath, strings.Join(op.Args, " "))
			fmt.Printf("exit code: %v\n", op.ExitCode)
			fmt.Printf("(stdin)\n%s\n", op.Stdin)
			fmt.Printf("(output)\n%s\n", strings.TrimSpace(op.Stdouterr))
		}

		if op.HaltIfFail && !op.Success {
			cleanExit("op failed and halt-if-fail is set")
		}
	}
}

func (r *Runner) DumpState(block *Block) {
	write := func(label string, body string) string {
		filename := fmt.Sprintf("%s/%05d.%s.%s", r.Config.StateDir, block.OpCounter, block.Label, label)
		err := os.WriteFile(filename, []byte(body), 0644)
		PanicIfErr(err)
		return filename
	}

	file := write("output", block.Stdouterr)
	block.StdouterrFile = file
	write("exitcode", fmt.Sprintf("%d", block.ExitCode))
	run, err := json.MarshalIndent(block, "", "  ")
	PanicIfErr(err)
	write("run.json", string(run)+"\n")
}

func cleanExit(msg string) {
	fmt.Println("EXITING - " + msg)
	os.Exit(1)
}
