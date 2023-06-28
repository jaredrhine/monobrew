package monobrew

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func ExecuteOp(op *Block) {
	op.CommandPath, _ = exec.LookPath(op.Command)
	op.StartTime = time.Now()
	exitcode, output, error := ExecuteCommand(op.Stdin, op.CommandPath, op.Args)
	op.EndTime = time.Now()
	op.ElapsedTime = op.EndTime.Sub(op.StartTime)
	op.RunError = error
	if len(output) == 0 {
		op.StdouterrIsEmpty = true
	}
	op.Stdouterr = output
	op.ExitCode = exitcode
	if exitcode == 0 {
		op.Success = true
	}
}

func ExecuteCommand(stdin string, command string, args []string) (exitCode int, output string, error string) {
	exitCode = 0

	cmd := exec.Cmd{}
	cmd.Path = command
	cmd.Args = append([]string{command}, args...)
	cmd.Stdin = strings.NewReader(stdin)
	cmd.Env = os.Environ()
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		errormsg := fmt.Sprintf("%s", err)
		if errormsg != "" {
			error = errormsg
		}
	}
	exitCode = cmd.ProcessState.ExitCode()

	output = string(stdoutStderr)

	return exitCode, output, error
}
