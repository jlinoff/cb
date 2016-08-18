package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// FileInfo is the file information associated with each line.
type FileInfo struct {
	fname   string
	base    string
	abspath string
	dir     string
}

// LineInfo is the line information. Used for generating error messages.
type LineInfo struct {
	fi     *FileInfo
	line   string
	lineno int
}

// RecipeStepType is the recipe const type.
type RecipeStepType int

const (
	stepCd RecipeStepType = iota
	stepExec
	stepExecNoExit
	stepExport
	stepInfo
	stepMustExistDir
	stepMustExistFile
	stepMustNotExistDir
	stepMustNotExistFile
	stepScript
)

// RecipeStep components.
type RecipeStep struct {
	Directive       RecipeStepType
	DirectiveString string
	Data            string
}

// RecipeInfo stores the information for a recipe.
type RecipeInfo struct {
	File      string
	Name      string
	Full      string
	Brief     string
	Aliases   map[string]int
	Variables map[string]string
	Steps     []RecipeStep
}

// runRecipe runs a recipe.
func runRecipe(opts CliOptions) {
	// We need to load the recipe to get the variable names.
	recipe := loadRecipe(opts.Recipe)

	// Set the recipe variables.
	runRecipeInitVariables(&recipe, opts)

	// Execute the steps.
	for i, step := range recipe.Steps {
		// Update the variables before each step.
		// This is done here to allow the variables to be changed dynamically.
		if strings.Contains(step.Data, "${") {
			for key, val := range recipe.Variables {
				variable := fmt.Sprintf("${%v}", key) // format is ${<name>}.
				n := strings.Replace(step.Data, variable, val, -1)
				if n != step.Data {
					step.Data = n
				}
			}
		}

		// Report step information.
		if strings.Contains(step.Data, "\n") {
			Log.Info("step.start = %v %v %v", i+1, step.DirectiveString, "multi-line")
		} else {
			Log.Info("step.start = %v %v %v", i+1, step.DirectiveString, step.Data)
		}
		wd, _ := os.Getwd()
		Log.Info("step.pwd = %v %v", i+1, wd)

		// Run the step.
		stepStart := time.Now()
		switch step.Directive {
		case stepCd:
			Chdir(step.Data)
		case stepExec:
			RunCmd("%v", step.Data)
		case stepExecNoExit:
			RunCmdNoExit("%v", step.Data)
		case stepExport:
			flds := strings.SplitN(step.Data, "=", 2)
			key := flds[0]
			val := flds[1]
			err := os.Setenv(key, val)
			if err != nil {
				Log.Err("failed to set the environment variable '%v' - %v", key, err)
			}
		case stepInfo:
			// Can't use Log.Info() here because the output will be lost if
			// --quiet is specified.
			Log.Printf("%v\n", step.Data)
		case stepMustExistDir:
			if IsDir(step.Data) == false {
				Log.Err("directory does not exist: %v", step.Data)
			}
		case stepMustExistFile:
			if IsFile(step.Data) == false {
				Log.Err("file does not exist: %v", step.Data)
			}
		case stepMustNotExistDir:
			if IsDir(step.Data) == true {
				Log.Err("directory exists: %v", step.Data)
			}
		case stepMustNotExistFile:
			if IsFile(step.Data) == true {
				Log.Err("file exists: %v", step.Data)
			}
		case stepScript:
			// Create a temporary script and execute it.
			// Note that it does not have to be executable.
			MkdirAll(Context.ScriptDir, 0700)
			fn := fmt.Sprintf("%v/%v.sh", Context.ScriptDir, Context.UserPID)
			Log.Info("creating anonymous script file: %v", fn)
			fp, err := os.Create(fn)
			if err != nil {
				Log.Err("can't create tmp file for script directive: %v", fn)
			}
			fmt.Fprintf(fp, "%v", step.Data)
			fp.Close()
			os.Chmod(fn, 0700)
			cmd := fn
			RunCmd(cmd)
			Log.Info("deleting anonymous script file: %v", fn)
			os.Remove(fn)
		default:
			Log.Err("unrecognized directive %v (%v) in %v", step.Directive, step.DirectiveString, recipe.File)
		}
		Log.Info("step.end = %v %.03f", i+1, time.Since(stepStart).Seconds())
	}
}

// runRecipeInitVariables initializes the recipe variables.
func runRecipeInitVariables(recipe *RecipeInfo, opts CliOptions) {
	// Use the recipe variable names to check the extra arguments.
	// Convert the arguments to options.
	ropts := map[string]string{}
	ks := []string{}
	for k := range recipe.Variables {
		o := "--" + k
		ropts[o] = k
		ks = append(ks, o)
	}

	// Check the options.
	for i := 0; i < len(opts.ExtraArgs); i++ {
		opt := opts.ExtraArgs[i]

		// Make sure that the option is valid.
		_, ok := ropts[opt]
		if ok == false {
			if len(ks) > 0 {
				sort.Strings(ks)
				Log.Err("invalid option specified '%v', valid options are %v", opt, ks)
			} else {
				Log.Err("invalid option specified '%v', there are no valid options", opt)
			}
		}

		// Now get the value.
		i++
		if i >= len(opts.ExtraArgs) {
			Log.Err("missing argument for '%v'", opt)
		}
		val := opts.ExtraArgs[i]
		key := ropts[opt]
		recipe.Variables[key] = val
	}

	// Verify that all of the required variables have values.
	en := 0
	for key, val := range recipe.Variables {
		if val == "" {
			en++
			Log.ErrNoExit("option '--%v' has no value", key)
		}
	}
	if en > 0 {
		Log.Err("unset variables found, cannot continue")
	}

	// Do the variable substitution for all variables.
	// The substitutions for the steps are done just-in-time to
	// allow the variable values to be updated dynamically.
	for key, val := range recipe.Variables {
		variable := fmt.Sprintf("${%v}", key)

		// Replace all occurrences of the variable in the value portion
		// of the variables. This does not affect the outer loop because
		// we are not changing the key.
		for n1, v1 := range recipe.Variables {
			if n1 != key { // skip ourself
				v2 := strings.Replace(v1, variable, val, -1)
				if v2 != v1 {
					recipe.Variables[n1] = v2
				}
			}
		}
	}
}

// loadRecipe loads a recipe.
func loadRecipe(recipeRef string) (recipe RecipeInfo) {
	if len(recipeRef) == 0 {
		Log.Err("null recipes not allowed")
	}
	Log.Info("loading recipe '%v'", recipeRef)
	recipeFile := ""
	if IsFile(recipeRef) {
		recipeFile = recipeRef
	} else {
		recipeFile = recipeRef

		// Add the INI extension.
		// This will catch things like ./foo --> ./foo.ini
		if strings.HasSuffix(recipeFile, ".ini") == false {
			Log.Info("appending extension: '.ini'")
			recipeFile = fmt.Sprintf("%v.ini", recipeFile)
		}

		// Prepend the recipes directory structure.
		if recipeFile[0] != '/' {
			Log.Info("prepending directory path: '%v'", Context.RecipeDir)
			recipeFile = path.Join(Context.RecipeDir, recipeFile)
		}
	}
	Log.Info("recipe file '%v'", recipeFile)
	nested := map[string]int{}
	lines := readRecipeFile(recipeFile, nested)
	recipe = makeRecipe(recipeFile, lines)
	return
}

// listAllRecipes lists all of the recipes with their brief descriptions.
func listAllRecipes() {
	recipes := loadAllRecipes()

	// Get the maximum width for the name field to allign all of the brief
	// descriptions.
	m := 0
	for _, recipe := range recipes {
		if len(recipe.Name) > m {
			m = len(recipe.Name)
		}
	}

	// Print out the data.
	for _, recipe := range recipes {
		fmt.Printf("%-*s - %v\n", m, recipe.Name, recipe.Brief)
	}
}

// loadAllRecipes loads all of the recipes in the path
func loadAllRecipes() (recipes []RecipeInfo) {
	files, e := ioutil.ReadDir(Context.RecipeDir)
	if e != nil {
		Log.Err("cannot read recipes directory: %v - %v", Context.RecipeDir, e)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		recipeFile := path.Join(Context.RecipeDir, file.Name())
		if strings.HasSuffix(recipeFile, ".ini") {
			// .ini files are recipe files.
			nested := map[string]int{}
			lines := readRecipeFile(recipeFile, nested)
			recipe := makeRecipe(recipeFile, lines)
			recipes = append(recipes, recipe)
		}
	}
	return
}

// readRecipeFile reads a file line by line and returns all of the lines.
func readRecipeFile(fname string, nested map[string]int) (lines []LineInfo) {
	if _, e := os.Stat(fname); os.IsNotExist(e) {
		Log.Err("recipe file does not exist: %v", fname)
	}

	// Check for nested references that could lead to infinite recursion.
	a, e := filepath.Abs(fname)
	if e != nil {
		Log.Err("cannot get abspath for recipe %v", a)
	}
	if _, ok := nested[a]; ok {
		// nested include found, report the stack
		i := 0
		for k := range nested {
			Log.Info("nested %3d %v", i, k)
			i++
		}
		Log.Err("nested include found - infinite recursion: %v\n", a)
	}
	nested[a] = 1

	// Cache the file information and open it for reading.
	fi := FileInfo{fname: fname, base: filepath.Base(a), abspath: a, dir: filepath.Dir(a)}
	fp, e := os.Open(fname)
	if e != nil {
		return
	}
	defer fp.Close()

	// Scan the file contents.
	s := bufio.NewScanner(fp)
	for lineno := 1; s.Scan(); lineno++ {
		line := s.Text()
		nextLineno := lineno

		// Look for include <file> statements.
		x := strings.TrimSpace(line)
		if strings.HasPrefix(x, "include") {
			// convert byte to rune for WS check
			r, _ := utf8.DecodeRune([]byte{x[7]})
			if unicode.IsSpace(r) {
				ifn := strings.TrimSpace(x[7:])
				if ifn[0] == '"' {
					ifn, e = strconv.Unquote(ifn)
					if e != nil {
						panic(e)
					}
				}
				if ifn[0] != '/' {
					// If the include reference is not an absolute path, use the
					// the path to the original file as the base directory.
					ifn = path.Join(fi.dir, ifn)
				}
				ilines := readRecipeFile(ifn, nested)
				lines = append(lines, ilines...)
				continue
			} else {
				// Syntax error.
				Log.Err("syntax error at include statement at line %v in %v", lineno, fi.abspath)
				continue
			}
		} else if len(x) == 0 || x[0] == '#' {
			// Skip blank lines and comments.
			continue
		} else {
			// Look for the special case of multi-line strings.
			//
			// They are declared as
			//    foo = """
			//    line1
			//    line2
			//    .
			//    .
			//    etc.
			//    """
			re1 := regexp.MustCompile(`^\S+\s*=\s*"""`)
			re2 := regexp.MustCompile(`"""$`)
			re3 := regexp.MustCompile(`^(\S+\s*=)\s*"""(.+)"""\s*$`)
			re4 := regexp.MustCompile(`^\S+\s*=\s*script\s+"""\s*(.*)$`)
			re5 := regexp.MustCompile(`^\S+\s*=\s*info\s+"""\s*(.*)$`)
			if re3.MatchString(x) {
				// It is all on a single line.
				// Example:
				//    foo = """ spam """
				m := re3.FindAllStringSubmatch(x, -1)
				x1 := strings.TrimSpace(m[0][1])
				x2 := strings.TrimSpace(m[0][2])
				line = fmt.Sprintf("%v %v", x1, x2)
			} else if re1.MatchString(x) {
				// Now parse until the end of the string.
				f := false
				for ; s.Scan(); lineno++ {
					sline := s.Text()
					line += "\n"
					line += sline
					x3 := strings.TrimSpace(sline)
					if re2.MatchString(x3) {
						nextLineno = lineno
						f = true
						break
					}
				}
				if f == false {
					Log.Err("syntax error: end of multiline string not found, starts at line %v in %v", lineno, fi.abspath)
				}
			} else if re4.MatchString(x) || re5.MatchString(x) {
				// Now parse until the end of the string.
				// Only for script and info.
				f := false
				for ; s.Scan(); lineno++ {
					sline := s.Text()
					line += "\n"
					line += sline
					x3 := strings.TrimSpace(sline)
					if re2.MatchString(x3) {
						nextLineno = lineno
						f = true
						break
					}
				}
				if f == false {
					Log.Err("syntax error: end of multiline string not found, starts at line %v in %v", lineno, fi.abspath)
				}
			}
		}

		// Keep track of each line.
		lines = append(lines, LineInfo{fi: &fi, lineno: lineno, line: line})
		lineno = nextLineno // account for multi-line strings
	}
	delete(nested, a)
	return
}

// makeRecipe makes the recipe object.
func makeRecipe(recipeFile string, lines []LineInfo) (rec RecipeInfo) {
	// Populate the recipe with initial values.
	n, a := getRecipeName(recipeFile)
	rec = RecipeInfo{
		Name:      n,
		File:      a,
		Variables: map[string]string{},
		Steps:     []RecipeStep{}}

	// Verify that no statements exist outside of a section.
	checkValidSections(recipeFile, lines)

	// valid step directives
	validStepDirective := map[string]RecipeStepType{
		"cd":                  stepCd,
		"export":              stepExport,
		"exec":                stepExec,
		"exec-no-exit":        stepExecNoExit,
		"info":                stepInfo,
		"must-exist-dir":      stepMustExistDir,
		"must-exist-file":     stepMustExistFile,
		"must-not-exist-dir":  stepMustNotExistDir,
		"must-not-exist-file": stepMustNotExistFile,
		"script":              stepScript,
	}

	// at this point we know that the syntax is sound so we need
	// to parse it into the recipe data structure for execution
	section := ""
	re1 := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z_\-0-9]*$`)
	re2 := regexp.MustCompile(`(?s)^(\S+)\s+(\S.*)$`) // handle multiline
	for _, li := range lines {
		line := li.line
		if line[0] == '[' {
			section = line // save the session name for context, later
			continue
		}

		// Get the key/value pairs.
		key, value := getRecipeAssignmentValue(li)

		// Assign them.
		switch section {
		case "[description]":
			switch key {
			case "brief":
				rec.Brief = value
			case "full":
				rec.Full = value
			}
		case "[variable]":
			if re1.MatchString(key) {
				rec.Variables[key] = value
			} else {
				Log.Err("invalid variable name '%v' at line %v in %v", key, li.lineno, li.fi.abspath)
			}
		case "[step]":
			// For a step we determine the directive, verify that it is valid
			// and then capture the rest of the line.
			m := re2.FindAllStringSubmatch(value, -1)
			directive := m[0][1]
			value = strings.TrimSpace(m[0][2])
			stype, ok := validStepDirective[directive]
			if ok == false {
				Log.Err("unknown step directive '%v' at line %v in %v", directive, li.lineno, li.fi.abspath)
			}
			rec.Steps = append(rec.Steps, RecipeStep{Directive: stype, DirectiveString: directive, Data: value})
			if stype == stepExport {
				// Export has a specific syntax, check it.
				if strings.Contains(value, "=") == false {
					Log.Err("export is of the form VAR=VAL, could not find '=' at line %v in %v", li.lineno, li.fi.abspath)
				}
			}
			break
		default:
			Log.Err("unknown section '%v' at line %v in %v", section, li.lineno, li.fi.abspath)
			break
		}
	}

	// Check the recipe to make sure that it has required fields set.
	if len(rec.Brief) == 0 {
		Log.Err("[description] brief not set for %v", rec.Name)
	}
	if len(rec.Full) == 0 {
		Log.Err("[description] full not set for %v", rec.Name)
	}
	if len(rec.Steps) == 0 {
		Log.Err("no steps defined in the [step] section for %v", rec.Name)
	}
	return
}

// getRecipeAssignmentValue gets the value associated with an assignment.
// This can be tricky for multiline strings for full and scripts.
func getRecipeAssignmentValue(li LineInfo) (key string, value string) {
	line := li.line
	re1 := regexp.MustCompile(`(?s)^\s*script\s+"""(.*)+"""$`)
	re2 := regexp.MustCompile(`(?s)^\s*info\s+"""(.*)+"""$`)
	re3 := regexp.MustCompile(`(?s)^\s*info\s+"`)

	// Get the key/value pairs.
	tokens := strings.SplitN(line, "=", 2)
	key = strings.TrimSpace(tokens[0])
	value = strings.TrimSpace(tokens[1])
	if strings.HasPrefix(value, `"""`) {
		// Handle lines of the form:
		//    <key> = """
		//    """
		value = value[3 : len(value)-3]
	} else if len(value) > 0 && value[0] == '"' {
		// Handle lines with quotes like this:
		//   <key> = "<value>"
		var e error
		value, e = strconv.Unquote(value)
		if e != nil {
			Log.Err("internal error, unquote operation failed at line %v in %v", li.lineno, li.fi.abspath)
		}
	} else if re1.MatchString(value) {
		// Handle lines of the form:
		//   step = script """#!/bin/bash
		//   """
		m := re1.FindAllStringSubmatch(value, -1)
		if len(m) > 0 {
			value = "script " + strings.TrimSpace(m[0][1])
		}
	} else if re2.MatchString(value) {
		// Handle lines of the form:
		//   step = info """
		//   stuff
		//   """
		m := re2.FindAllStringSubmatch(value, -1)
		if len(m) > 0 {
			value = "info " + strings.TrimSpace(m[0][1])
		}
	} else if re3.MatchString(value) {
		// Handle lines of the form:
		//   step = info "this is a test"  --> this is a test
		p := strings.Index(value, `"`)
		s := value[p:]
		u, e := strconv.Unquote(s)
		if e != nil {
			Log.Err("internal error, unquote operation failed at line %v in %v", li.lineno, li.fi.abspath)
		}
		value = "info " + u
	}
	return
}

// getRecipeName gets the recipe name
func getRecipeName(recipeFile string) (name string, abspath string) {
	// Get the name.
	fn := filepath.Base(recipeFile)
	ext := filepath.Ext(fn)
	name = fn[:len(fn)-len(ext)]
	a, err := filepath.Abs(recipeFile)
	if err != nil {
		Log.Err("invalid file path: %v", err)
	}
	abspath = a
	return
}

// checkValidSections
func checkValidSections(recipeFile string, lines []LineInfo) {
	// valid sections and decl keywords within the section
	validSections := map[string]map[string]int{
		"[description]": {"brief": 0, "full": 0},
		"[variable]":    {},
		"[step]":        {"step": 0}}

	// Verify that no statements exist outside of a section.
	section := ""
	for _, li := range lines {
		line := li.line
		if line[0] == '[' {
			// This is a section, see if it is a valid one.
			if _, ok := validSections[line]; ok == false {
				Log.Err("invalid section found: %v at line %v in %v", line, li.lineno, li.fi.abspath)
			}
			section = line
			continue
		}

		// All statements must be inside a section, check for orphans.
		if section == "" {
			Log.Err("orphan declaration at line %v in %v: %v", li.lineno, li.fi.abspath, li.line)
		}

		// Check for an equals sign.
		if strings.Contains(line, "=") == false {
			Log.Err("syntax error, missing '=' at line %v in %v: %v", li.lineno, li.fi.abspath, li.line)
		}

		// Make sure that the tokens are valid.
		tokens := strings.SplitN(line, "=", 2)
		decl := strings.TrimSpace(tokens[0])
		if section != "[variable]" {
			if _, ok := validSections[section][decl]; ok == false {
				Log.Err("syntax error, found invalid declaration '%v' in section '%v' at line %v in %d: %v", decl, section, li.lineno, li.fi.abspath, li.line)
			}
			validSections[section][decl] = 1
		}
	}
}
