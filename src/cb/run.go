// The run target.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jlinoff/go/run"
)

// RunOpt runs command from the --run option.
// Example:
//      --run /bin/bash -c 'echo -e "this is a \ntest"'
func RunOpt(opts []string) {
	Log.Info("cmd    = %v", MakeCmdString(opts))
	Log.Info("method = run.Cmd")
	s := time.Now()
	stdout, err := run.Cmd(opts)
	Log.Info("time   = %.03f", time.Since(s).Seconds())
	Log.Info("size   = %v", len(stdout))
	if err == nil {
		Log.Info("status = passed")
	} else {
		Log.Info("status = failed - %v", err)
	}
	return
}

// RunSilentOpt runs command from the --run-silent option.
// Example:
//      --run /bin/bash -c 'echo -e "this is a \ntest"'
func RunSilentOpt(opts []string) {
	Log.Info("cmd    = %v", MakeCmdString(opts))
	Log.Info("method = run.Cmd")
	s := time.Now()
	stdout, err := run.CmdSilent(opts)
	Log.Info("time   = %.03f", time.Since(s).Seconds())
	Log.Info("size   = %v", len(stdout))
	if err == nil {
		Log.Info("status = passed")
	} else {
		Log.Info("status = failed - %v", err)
	}
	return
}

// RunCmd runs a command with logging.
// It exits if an error occurred.
// It is a wrapper for runCmd.
func RunCmd(f string, a ...interface{}) (err error) {
	cmd := fmt.Sprintf(f, a...)
	wd, _ := os.Getwd()
	Log.InfoWithLevel(3, "cmd.cmd = %v", cmd)
	Log.InfoWithLevel(3, "cmd.pwd = %v", wd)
	s := time.Now()
	err = run.CmdWithWriters(TokenizeString(cmd), Log.Writers)
	Log.InfoWithLevel(3, "cmd.elapsed = %.03f", time.Since(s).Seconds())
	if err == nil {
		Log.InfoWithLevel(3, "cmd.status = passed")
	} else {
		code := run.GetExitCode(err)
		Log.ErrWithLevel(3, "cmd.status = failed (%v) - %v", code, err)
	}
	return
}

// RunCmdNoExit runs a command with logging.
// It does not exit if an error occurred.
// It is a wrapper for runCmd.
func RunCmdNoExit(f string, a ...interface{}) (err error) {
	cmd := fmt.Sprintf(f, a...)
	wd, _ := os.Getwd()
	Log.InfoWithLevel(3, "cmd.cmd = %v", cmd)
	Log.InfoWithLevel(3, "cmd.pwd = %v", wd)
	s := time.Now()
	err = run.CmdWithWriters(TokenizeString(cmd), Log.Writers)
	Log.InfoWithLevel(3, "cmd.elapsed = %.03f", time.Since(s).Seconds())
	if err == nil {
		Log.InfoWithLevel(3, "cmd.status = passed")
	} else {
		code := run.GetExitCode(err)
		Log.WarnWithLevel(3, "cmd.status = failed (%v) - %v", code, err)
	}
	return
}

// RunCmdSilent runs a command with logging.
// It exits if an error occurred.
// It is a wrapper for runCmd.
func RunCmdSilent(f string, a ...interface{}) (out string, err error) {
	cmd := fmt.Sprintf(f, a...)
	wd, _ := os.Getwd()
	Log.InfoWithLevel(3, "cmd.cmd = %v", cmd)
	Log.InfoWithLevel(3, "cmd.pwd = %v", wd)
	s := time.Now()
	out, err = run.CmdSilent(TokenizeString(cmd))
	Log.InfoWithLevel(3, "cmd.elapsed = %.03f", time.Since(s).Seconds())
	if err == nil {
		Log.InfoWithLevel(3, "cmd.status = passed")
	} else {
		code := run.GetExitCode(err)
		Log.ErrWithLevel(3, "cmd.status = failed (%v) - %v", code, err)
	}
	return
}
