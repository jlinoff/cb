// See the online help (-h) for detailed information.
// To learn how to add a target see the documentation in target.go.
package main

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/jlinoff/go/msg"
)

// Log is the package logger. It writes to os.Stdout and, optionally, a file
// if -t is specified.
var Log *msg.Object

func main() {
	opts := NewCliOptions()
	if opts.Quiet {
		Log.DebugEnabled = false
		Log.InfoEnabled = false
		Log.WarningEnabled = false
	}

	// Create the tee file, if necessary.
	tee := ""
	if opts.Tee {
		t := time.Now().Truncate(time.Millisecond).Format("20060102-150505")
		b := path.Base(os.Args[0])
		u, _ := user.Current()
		tee = fmt.Sprintf("%v-%v-%v.log", b, t, u.Username)
		Log.Info("tee file : %v", tee)
		fp, err := os.Create(tee)
		if err != nil {
			Log.Err("unable to open log file: %v", tee)
		}
		Log.Writers = []io.Writer{os.Stdout, fp}
		defer fp.Close()
	}

	// Display the run-time context.
	MakeContext(tee)
	if opts.RecipeDir != "" {
		Context.RecipeDir = opts.RecipeDir
	}
	Context.PrintContext()

	// Define the pre-defined environment variables.
	// Defining them here makes them available for recipes, help and
	// run operations.
	setenv := func(name string, val string) {
		vname := strings.ToUpper(fmt.Sprintf("%v_%v", Context.Base, name))
		Log.Info("built-in env var %v=%v", vname, val)
		os.Setenv(vname, val)
	}
	setenv("base", Context.Base)
	setenv("builddate", Context.MakeBuildDate)
	setenv("exe", Context.ExePath)
	setenv("pid", strconv.Itoa(Context.UserPID))
	setenv("pwd", Context.Pwd)
	setenv("recipes", Context.RecipeDir)
	setenv("scripts", Context.ScriptDir)
	setenv("timestamp", Context.TimeStamp)
	setenv("username", Context.UserName)
	setenv("version", Context.MakeVersion)

	// Perform the specified action.
	switch opts.Action {
	case actionHelp:
		help(opts)
	case actionList:
		listAllRecipes()
	case actionRecipe:
		runRecipe(opts)
	case actionRun:
		RunOpt(opts.ExtraArgs)
	case actionRunSilent:
		RunSilentOpt(opts.ExtraArgs)
	default:
		break
	}
}

// init the logger
func init() {
	n := path.Base(os.Args[0])
	f := `%pkg %(-27)time %(-7)type %(-7)file %(4)line - %msg`
	t := `2006-01-02 15:05:05.000 MST`
	w := []io.Writer{os.Stdout}
	l, e := msg.NewMsg(n, f, t, w)
	if e != nil {
		fmt.Fprintf(os.Stderr, "ERROR: logger creation failed: %v", e)
		os.Exit(1)
	}
	Log = l
}
