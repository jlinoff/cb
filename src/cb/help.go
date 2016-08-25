package main

// help prints the help message and exits.
import (
	"fmt"
	"os"
	"strings"
)

//  help generates the help message.
func help(opts CliOptions) {
	if opts.HelpArg == "" {
		helpTop()
	} else {
		// generate the help for a recipe
		recipe := loadRecipe(opts.HelpArg)
		fmt.Printf("Help for %v - %v\n", recipe.Name, recipe.File)
		fmt.Printf("%v\n", recipe.Full)
	}
	os.Exit(0)
}

// helpTop generates the top level help.
func helpTop() {
	msg := `
USAGE
    %[1]v [OPTIONS] <RECIPE> [RECIPE_OPTIONS]

DESCRIPTION
    %[1]v runs recipes that perform complex build tasks that normally require
    lots of different steps or arcane combinations of command line options.

    Sounds a lot like scripts, right? Why not just create scripts to wrap the
    complex commands. You can and should use scripts to wrap complex functions.
    Scripts are great.

    However, when there is a profileration of many scripts for a class of tasks
    (like building SW), it is sometimes difficult to find the one that you want.

    That is where recipes come in. They are not meant to replace scripts or
    other tools. Instead they are meant to help organize them in a way that
    makes it easy to find the functionality that you need.

    Recipes are simply wrappers for steps that live in a central place with
    brief and full descriptions to make it easier to find them and simple
    support variables to allow them to be customized a bit.

    Recipes can be created by anyone and shared with everyone or kept private.

    The syntax for using recipes is very simple:
        %[1]v [OPTIONS] <recipe> [RECIPE_OPTIONS]

    Here is how you list all recipes along with a brief description:
        $ %[1]v --list

    Here is how you get detailed information for a specific recipe:
        $ %[1]v help <recipe>

    Here is how you run a recipe:
        $ %[1]v <recipe> <recipe-options>

    Each recipe is defined by an INI file with three parts: the description,
    the variables and the steps. These sections are described in detail in
    the next section.

MOTIVATION
    The motivation for developing this tool came from working with a legacy
    build system that had many tasks and each task required a dozen or more
    steps. This first approach was to write scripts to wrap the steps but that
    quickly got out of hand with dozens of them so this approach was developed
    which made things much easier. It also hid the implementation so that it
    could be improved.

RECIPES
    Recipes are the heart of the system. Each recipe defines a sequence of
    commands for performing an operation.

    Recipes are defined by INI files with a .ini extension.

    Blank lines are ignored.

    Lines that start with a '#' as the first non-whitespace character are
    comment lines that are ignored.

    Multi-line strings can be specified by delimiting them with """.

    Includes are allowed. An include is defined by they keyword 'include'
    followed by a filename. Include statements can appear anywhere in the
    file. They act just like paste operations and can be used to share code.

    Include files can include other files. Include files must not have a .ini
    extension. The recommended extension is .inc but anything will work.

    Recipes have three sections:

        [description]  Fields that describe the recipe.
        [variable]     Defines variables for the recipe.
        [steo]         Defines the recipe steps.

    The description section contains two variable: brief and full. Brief is a
    one line description of the recipe. Full is a full multiline description.
    You can use """ """ syntax for the full description.

    The variable section defines variables that the user can change. Each
    variable has a name and an optional value separated by an equals '=' sign.

    Variable names can only be composed of lower case letters, digits,
    underscores and dashes. They cannot start with a digit or a dash.

    Variables are referenced using shell syntax: ${name}. Note that the braces
    are not optional. If the variables are assigned a value, that is the default
    value. If they are not assigned a value, then they are required.

    Variable names appear as options on the command line. That means that
    if you define a variable named "foo", an option named --foo will be
    generated to set that variable. Here are some sample declarations of
    variables to make this clearer.

        # defines two variables, required and optional.
        # they appear as --required <value> and --optional <value> on the
        # command line.
        [variable]
        required =
        option = default

    The step section defines the steps taken. It is very simple and does not
    support looping or conditionals. That is because it is only meant to handle
    high level operations that deal with running multiple scripts in order. For
    lower level stuff that requires looping or conditionals, it makes more sense
    to use a script. Note that you can embed an anonymous script if you don't
    want to create an external one explicitly.

    Each entry in the step section is defined like this:

        step = <directive> <data>

    The directive tells %[1]v what to do. The following directives are
    available.

        cd <dir>                    Change the working dir for all subsequent steps.

        export X=Y                  Define an env var for all subsequent steps.

        exec <cmd>                  Execute a command, stop if it fails.

        exec-no-exit <cmd>          Exexute a command, continue if it fails.

        info <msg>                  Print a message to the log.

        must-exist-dir <dir>        Fail if <dir> does not exist.
                                    Shorthand for
                                    step = exec /bin/bash -c "[ -d <dir>] && exit 0 || exit 1"

        must-exist-file <file>      Fail if <file> does not exist.
                                    Shorthand for
                                    step = exec /bin/bash -c "[ -f <file>] && exit 0 || exit 1"

        must-not-exist-dir <dir>    Fail if <dir> exists.
                                    Shorthand for
                                    step = exec /bin/bash -c "[ ! -d <dir>] && exit 0 || exit 1"

        must-not-exist-file <file>  Fail if <file> exists.
                                    Shorthand for
                                    step = exec /bin/bash -c "[ ! -f <file>] && exit 0 || exit 1"

        script                      Embed an anonymous, in-line script.
                                    You can use any scripting language.
                                    They are generated dynamically in %[3]v.
                                    You can change a variable setting by
                                    specifying a line of the form:
                                        ###export <variable> = <value>

    Here is an example recipe. It is named list-files.ini so you can refer to it
    as "list-files" on the command line.

        # This is an example recipe.
        [description]
        brief = "List files in a directory."
        full = """
        USAGE
            list-files [OPTIONS]

        DESCRIPTION
            Lists files in a directory.

        OPTIONS
            --dir DIR    Override the default directory.
        """

        [variable]
        dir = /tmp

        [step]
        step = must-exist-dir ${dir}

        step = info "ls command"
        step = ls -l ${dir}

        step = info """
        # ================================================================
        # run anonymous bash and python scripts
        # ================================================================
        """

        step = script """#!/bin/bash
        echo "bash script - ${dir}"
        """

        step = script """#!/usr/bin/env python
        print("python script - {}".format("${dir}"))
        """

        # Reset the value of the dir variable from an anonymous script.
        # It will be /var for all subsequent steps.
        step = script """#!/bin/bash
        echo "###export dir=/var"
        """

        # dir will be /var
        step = exec ls -l ${dir}

        step = info done

    Note the use of '"""' to delimit multi-line strings for the full description
    and the anonymous script.

    You would run this recipe like this:
        $ %[1]v list-files

    To print the help, do this:
        $ %[1]v help list-files

    To list a different directory, do this:
        $ %[1]v list-files --dir /var/run

ENVIRONMENT VARIABLES
    When a recipe is run there are environment variables that are made available
    to it by %[1]v. The list of environment variables is shown below.

        %[4]v

    To use an environment variable just reference it like a normal variable.
    Here is an example: ${%[2]v_USERNAME}.

CALLING OTHER RECIPES
    You can use ${%[2]v_EXE} to call other recipes like this:

        # Call other another recipe.
        step = exec ${%[2]v_EXE} --arg1 arg1

    Use this approach with caution because you could end up with infinite
    recursion for a recipe that calls itself.

OPTIONS
    -h, --help         On-line help. Same as "%[1]v help".

    -f FILE, --flatten FILE
                       Flatten a recipe into a file.

    -l, --list         List the available recipes with a brief description.

    -q, --quiet        Run quietly. Only error messages are printed.
                       If -q and -v are not specified, only ERROR and WARNING
                       messages are printed.

    --no-banner        Turn off the step banner in verbose mode.

    -r DIR, --recipes DIR
                       The path to the recipes directory.

    --run <cmd> <args> Run a command. Used for internal testing.

    -t, --tee          Log all messages to a unique log file as well as stdout.
                       It saves having to create a unique file name for each run
                       using the command line tee tool.

                       The output file name is
                           %[1]v-<YYYYMMDD>-<hhmmss>-<username>.log

    -v, --verbose      Increase the level of verbosity.
                       It can be specified multiple times.
                           -v     --> print INFO and banner messages
                           -v -v  --> print INFO, banner and DEBUG messages
                       You always want to use -v when running recipes.

    -V, --version      Print the program version and exit.

EXAMPLES
    $ # Example 1: Get help.
    $ %[1]v help

    $ # Example 2: List all available recipes.
    $ %[1]v --list

    $ # Example 3: Get help about a recipe.
    $ %[1]v help <recipe>

    $ # Example 4: Show your local configuration.
    $ %[1]v -v

    $ # Example 5: Run a recipe with automatic logging.
    $ #            Provide a recipe specific option. The options
    $ #            are different for each recipe.
    $ %[1]v -v -t <recipe> --foo bar

    $ # Example 6: Run a local recipe file.
    $ %[1]v -v ./myrecipe.ini

    $ # Example 7: Use a local recipe repository.
    $ %[1]v -v -r ~/my/recipes myrecipe1

`
	// Get the built-in environment variables.
	evs := []string{}
	ub := strings.ToUpper(Context.Base)
	ubu := ub + "_"
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, ubu) {
			evs = append(evs, e)
		}
	}

	// Print the message and exit.
	fmt.Printf(msg, Context.Base, ub, Context.ScriptDir, strings.Join(evs, "\n        "))
	os.Exit(0)
}
