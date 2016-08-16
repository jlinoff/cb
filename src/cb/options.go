package main

import (
	"fmt"
	"os"
	"path"
)

// CliOptionsType defines the type of action.
type CliOptionsType int

const (
	actionUnknown CliOptionsType = iota
	actionHelp
	actionRecipe
	actionRun
	actionRunSilent
	actionList
)

// CliOptions are the command line options.
type CliOptions struct {
	Action      CliOptionsType
	HelpArg     string
	Tee         bool
	Quiet       bool
	ShellScript string
	Recipe      string
	RecipeDir   string
	ExtraArgs   []string
}

// NewCliOptions gets the command line options and figures out the
// action to take.
func NewCliOptions() (opts CliOptions) {
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch arg {
		case "-h", "--help", "help":
			// help can have an optional argument
			// can't do it here because the optional
			// argument is a recipe that requires searching
			// to find.
			opts.Action = actionHelp // overrides all other actions
			i++
			if i < len(os.Args) {
				opts.HelpArg = os.Args[i]
			}
		case "-l", "--list":
			// list means list all of the recipes along with their brief descriptions
			if opts.Action == actionUnknown {
				opts.Action = actionList // do not override other actions
			}
		case "-r", "--recipes":
			d := cliGetNextArg(&i)
			if IsDir(d) == false {
				Log.Err("not a valid directory: %v", d)
			}
			opts.RecipeDir = d
		case "--run":
			if opts.Action == actionUnknown {
				opts.Action = actionRun // do not override other actions
			}
			i++
			if i < len(os.Args) {
				opts.ExtraArgs = os.Args[i:]
			} else {
				Log.Err("missing arguments for --run")
			}
			i = len(os.Args)
		case "--run-silent":
			if opts.Action == actionUnknown {
				opts.Action = actionRunSilent // do not override other actions
			}
			i++
			if i < len(os.Args) {
				opts.ExtraArgs = os.Args[i:]
			} else {
				Log.Err("missing arguments for --run")
			}
			i = len(os.Args)
		case "-s", "--shell":
			// generate a shell script
			opts.ShellScript = cliGetNextArg(&i)
		case "-t", "--tee":
			// tee the output to a unique file name
			opts.Tee = true
		case "-q", "--quiet":
			// run quietly
			opts.Quiet = true
		case "-V", "--version":
			base := path.Base(os.Args[0])
			fmt.Printf("%v - v%v\n", base, Version)
			os.Exit(0)
		default:
			if arg[0] == '-' {
				Log.Err("unrecognized option '%v', try -h for more information", arg)
			}
			if opts.Action == actionUnknown {
				opts.Action = actionRecipe // do not override other actions
			}
			opts.Recipe = arg
			i++ // get the remaing arguments
			if i < len(os.Args) {
				opts.ExtraArgs = os.Args[i:]
			}
			i = len(os.Args)
		}
	}
	return
}

// cliGetNextArg gets the next command line argument.
func cliGetNextArg(i *int) (arg string) {
	opt := os.Args[*i]
	*i++
	if *i >= len(os.Args) {
		Log.Err("missing argument for option %v", opt)
	}
	return os.Args[*i]
}
