# cb
Cb (cookbook) runs recipes that encapsulate complex tasks composed of CLI tools from diverse sources.

## 1. Motivation

The motivation for developing this tool came from working with a legacy
build system that had many tasks and each task required a dozen or more
steps. This first approach was to write scripts to wrap the steps but that
quickly got out of hand with dozens of them so this approach was developed
which made things much easier. It also hid the implementation so that it
could be improved.

## 2. Introduction

This tool runs recipes that perform complex tasks that normally require lots of
different steps or arcane combinations of command line options.

Sounds a lot like scripts, right? Why not just create scripts to wrap the
complex commands. _You can and should_ use scripts to wrap complex functions.
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

The syntax for running a recipe is very simple:

    cb [OPTIONS] <recipe> [RECIPE_OPTIONS]

Here is how you list all recipes along with a brief description:

    $ cb -q --list

Here is how you get detailed information for a specific recipe:

    $ cb -q help <recipe>

Here is how you run a recipe:

    $ cb <recipe> <recipe-options>

Each recipe is defined by an INI file with three parts: the description,
the variables and the steps. These sections are described in detail in
the next section.

### 3. Installing
This is a go program. It has been tested with 1.6.3 on Mac OS X 10.11.6 and linux CentOS 7.2.

Here is how you would download and install it assuming the go 1.6.3 or later is in your path.

```bash
$ git clone https://github.com/jlinoff/cb.git
$ cd cb
$ make
$ bin/cb
$ bin/cb -q -h
$ bin/cb -q help test-script
```

## 4. Recipes
Recipes are the heart of the system. Each recipe defines a sequence of
commands for performing an operation.

Recipes are defined by INI files with a .ini extension.

Blank lines are ignored.

Lines that start with a '#' as the first non-whitespace character are
comment lines that are ignored.

Multi-line strings can be specified by delimiting them with `"""`.

Includes are allowed. An include is defined by they keyword 'include'
followed by a filename. Include statements can appear anywhere in the
file. They act just like paste operations and can be used to share code.

Include files can include other files.

Include files must not have a `.ini` extension. The recommended extension
is `.inc` but anything will work.

Recipes have three sections:

| Section       | Description |
| ------------- | ----------- |
| [description] | The full and brief fields that describe the recipe. |
| [variable]    | Defines variable for the recipe that can be changed at run time. |
| [step]        | Defines the recipe steps. |

### 4.1 [description]
The description section contains two variable: brief and full. Brief is a
one line description of the recipe. Full is a full multiline description.
You can use """ """ syntax for the full description.

Here is an example for a recipe named "awesome":

    [description]
    brief = "does awesome stuff"
    full = """
    USAGE
        awesome [OPTIONS]
        
    DESCRIPTION
        Does awesome stuff. Try it! You will like it.
        
    OPTIONS
        v1      The string to print.
        
    EXAMPLES
        awesome --v1 "print this string!"
    """

### 4.2 [variable]
The variable section defines variables that the user can change. Each
variable has a name and an optional value separated by an equals `=` sign.

Variable names can only be composed of lower case letters, digits,
underscores and dashes. They cannot start with a digit or a dash.

Variables are referenced using shell syntax: ${name}. Note that the braces
are not optional. If the variables are assigned a value, that is the default
value. If they are not assigned a value, then they are required.

Variable names appear as options on the command line. That means that
if you define a variable named "foo", an option named --foo will be
generated to set that variable.

Here are some examples of variable declarations to make this clearer.

    # defines two variables, required and optional.
    # they appear as --required <value> and --optional <value> on the
    # command line.
    [variable]
    required =
    option = default
    foo = bar
    spam = "spam is a ${foo}"
    multiline-var = """
    line 1
    line 2
    """

### 4.3 [step]
The step section defines the steps taken. It is very simple and does not
support looping or conditionals. That is because it is only meant to handle
high level operations that deal with running multiple scripts in order. For
lower level stuff that requires looping or conditionals, it makes more sense
to use a script. Note that you can embed an anonymous script for virtually
any scriptable language if you don't want to create an external one explicitly.

Each entry in the step section is defined like this:

    step = <directive> = <data>

Each directive is executed sequentially. The available directives are described
in the following table.

| Directive               | Description |
| ----------------------- | ----------- |
| cd DIR                  | Change the working directory to DIR. |
| export KEY=VALUE        | Define an environment variable that is accessible in all subsequent steps. |
| exec CMD                | Execute a command with optional arguments, stop if it fails. |
| exec-no-exit CMD        | Execute a command with optional arguments, do not stop if it fails. |
| info MSG                | Print a message to the log. |
| must-exist-dir DIR      | Fail if directory DIR does not exist.<br>This is shortand for<br>`step = exec /bin/bash -c "[ -d DIR ] && exit 0 || exit 1"` |
| must-exist-file FILE    | Fail if file FILE does not exist.<br>This is shortand for<br>`step = exec /bin/bash -c "[ -f FILE ] && exit 0 || exit 1"` |
| must-not-exist-dir DIR  | Fail if directory DIR exists.<br>This is shortand for<br>`step = exec /bin/bash -c "[ ! -d DIR ] && exit 0 || exit 1"` |
| must-not-exist-file FILE| Fail if file FILE exists.<br>This is shortand for<br>`step = exec /bin/bash -c "[ ! -f FILE ] && exit 0 || exit 1"` |
| script `""" ... """`      | Embed an anonymous, in-line script. You can use any scripting language. |

### 4.4 Setting variables from inside scripts

There are some occasions where you need to be able to change the
value of a recipe variable from inside a script to set state for subsequent
steps. Cb allows you to do this by recognizing specially formatted
messages in the output of the script. 

The message format for changing a variable value is shown below.

    ###export VARIABLE = VALUE
    
It must appear as a separate line.

If the variable does not exist, it will be created.

White space around the value is trimmed. If you want to keep white space,
you can quote the value.

### 4.5 Example Recipe
Here is a full example of a recipe.

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
    must-exist-dir ${dir}

    step = info "the ls command"
    step = ls -l ${dir}

    step = info """
    # ================================================================
    # run bash and python scripts
    # ================================================================
    """
    
    step = script """#!/bin/bash
    echo "bash script - ${dir}"
    """

    step = script """#!/usr/bin/env python
    print("python script - {}".format("${dir}"))
    """
    
    # Change dir to be /var for all subsequent steps.
    step = script """#!/bin/bash
    echo "###export dir=/var"
    """
    
    # Will list the contents of /var.
    step = exec ls -l ${dir}

    step = info done

## 5. Environment Variables

When a recipe is run the following are environment variables that are made available
to it by default.

| Env Var      | Description |
| ------------ | ----------- |
| CB_BASE      | Base name of package (CB). |
| CB_BUILDDATE | Date that the package was built. Set by the Makefile. |
| CB_PID       | Process ID of the job that is running the recipe. |
| CB_PWD       | The directory the command was started from. |
| CB_RECIPES   | The recipes directory. |
| CB_SCRIPTS   | The scripts cache directory. |
| CB_TIMESTAMP | The timestamp (suitable for use a file name) of the time that the run was started. |
| CB_USERNAME  | The username of the person running the recipe. |
| CB_VERSION   | The version of the tool, also set in the Makefile. |

## 6. CLI Options
The table below shows the CLI options that are available.

| Short<br>Option | Long<br>Option | Description   |
| --------------- | -------------- | ------------- |
| -f FILE         | --flatten FILE | Flatten a recipe into a file. Useful for debugging and dry run analyses. |
| -h              | --help         | Help message. |
| -l              | --list         | List the available recipes with a brief description. |
| -q              | --quiet        | Run quietly. Very useful for --help, --list and recipe help commands. |
| -r DIR          | -recipes DIR   | The path to the recipes directory. The default path ../etc/cb/recipes relative to the cb executable. |
| -t              | --tee          | Log all messages to a unique log file as well as stdout. It saves having to create a unique file name for each run using the command line tee tool. The format is cb-[YYYYMM]-[hhmms]-[USERNAME].log |
| -V              | --version      | Print the program name and exit. |

## 7. Examples

### 7.1 Get help.
```bash
$ cb -q -h
$ cb -q help
```

### 7.2 List all available recipes.
```bash
$ cb -q --list
```

### 7.3 Get help about a recipe named alice.
```bash
$ cb -q help alice
```

### 7.4 Show your local configuration.
```bash
$ cb
```

### 7.5 Run recipe bob with automatic logging.
```bash
$ cb -t bob
```

### 7.6 Run using local recipes.
```bash
$ cb -r ~/my/recipes myone
```

### 7.7 Run a recipe outside of a recipe archive.
```bash
$ cb /tmp/test-recipe.ini
```
