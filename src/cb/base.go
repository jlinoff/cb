// target type information and support routines
// These are the steps to create a new target:
//   1. choose the target name
//   2. create a go source file for it less than 8 characters in length with a
//      "t" prefix. If it absolutely must be longer than 8 characters, that is ok.
//   3. create a structure for the target that subclasses the TargetBase structure.
//   4. fill in the target name, information and brief description
//      the brief text is used in the help so it should be short
//      the info text is the help displayed when the user types "help <target>"
//   5. create the interface functions defined in the TargetInterface interface
//      and fill them in.
//   6. create the init() function to add the target to global Targets map when
//      the program starts up
package main

// Targets are the list of target objects available.
var Targets = map[string]TargetInterface{}

// TargetInterface is the interface that all targets must support.
type TargetInterface interface {
	Name() string
	Brief() string
	Description() string
	Help()
	Run([]string, CliOptions) (int, error)
	Aliases() []string
}

// TargetBase is the base of all targets.
type TargetBase struct {
	name    string
	info    string
	brief   string
	aliases []string
}

// getNextTargetArg - get next the argument from the target options.
func getNextTargetArg(i *int, opts []string, t string) (val string) {
	opt := opts[*i]
	*i++
	if len(opts) > *i {
		val = opts[*i]
	} else {
		Log.Err("missing argument to '%v' for target '%v'", opt, t)
	}
	return
}
