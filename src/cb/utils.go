package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"unicode"
)

// SandboxRoot get the directory that contains the repo. This is the top
// of the sandbox.
// If it is found, the sandbox directory is returned.
// If it is not found, the sandbox directory is "".
func SandboxRoot() (root string) {
	wd, _ := os.Getwd()
	ap, _ := filepath.Abs(wd)
	sep := fmt.Sprintf("%c", os.PathSeparator)
	d := strings.Split(ap, sep)
	for len(d) > 0 {
		repodir := sep + path.Join(d...)
		repo := path.Join(repodir, ".repo")
		if PathExists(repo) {
			root = repodir
			return
		}
		d = d[:len(d)-1]
	}
	return
}

// PathExists reports whether a path exists.
func PathExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// IsDir reports whether the path is a directory.
func IsDir(path string) bool {
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return fi.IsDir()
}

// IsFile reports whether the path is a file.
func IsFile(path string) bool {
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return fi.IsDir() == false
}

// MkdirAll makes all directories in the path.
func MkdirAll(path string, mode os.FileMode) {
	// If it already exists, don't bother.
	if IsDir(path) {
		return
	}
	// Make the directory if it does not exist.
	if err := os.MkdirAll(path, mode); err != nil {
		Log.ErrWithLevel(3, "mkdir operation failed for '%v': %v", path, err)
	}
}

// Chdir changes the directory.
func Chdir(path string) {
	Log.InfoWithLevel(3, "cd to %v", path)
	if err := os.Chdir(path); err != nil {
		Log.ErrWithLevel(3, "failed to change directory to %v", path)
	}
	wd, _ := os.Getwd()
	Log.InfoWithLevel(3, "pwd = %v", wd)
}

// GetDefaultRecipesDir gets the the default recipes directory.
func GetDefaultRecipesDir() (dir string) {
	a0, _ := filepath.Abs(os.Args[0])
	b := filepath.Base(a0)
	dir = path.Join(filepath.Dir(a0), "../etc", b, "recipes")
	return
}

// RepoProject contains the information for a repo project.
type RepoProject struct {
	name    string
	path    string
	abspath string
}

// GetRepoProjects gets the repo projects from the manifest.xml file.
func GetRepoProjects() (projects []RepoProject) {
	// Get the root of the sandbox.
	root := SandboxRoot()
	if root == "" {
		Log.ErrWithLevel(3, "not in a sandbox")
	}

	// Get the manifest.
	manifest := path.Join(root, ".repo", "manifest.xml")
	if _, err := os.Stat(manifest); os.IsNotExist(err) {
		Log.ErrWithLevel(3, "manifest file does not exist: %v", manifest)
	}

	// Read the manifest to get the project data.
	// It is an XML file of the form:
	//    <?xml version="1.0" encoding="UTF-8"?>
	//    <manifest>
	//       <remote name="origin" fetch="http://sfxsv-cmgerrit:8080" review="sfxsv-cmgerrit:8080" />
	//       <default remote="origin" revision="1.4.0" />
	//       <project name="cryptomanager" path="." />
	//       <project name="tools" path="tools" />
	//       <project name="cmservice" path="./devices/cmservice" />
	//       <project name="vmware-automation" path="vmware-automation" />
	//    </manifest>
	// Note that I am not using a schema here, I am only looking for the "project"
	// element.
	fp, err := os.Open(manifest)
	if err != nil {
		Log.ErrWithLevel(3, "could not read file: %v", manifest)
	}
	defer fp.Close()
	decoder := xml.NewDecoder(fp)
	for {
		token, _ := decoder.Token()
		if token == nil {
			break
		}
		switch elem := token.(type) {
		case xml.StartElement:
			if elem.Name.Local == "project" {
				// Grab the name and path attributes.
				n := ""
				p := ""
				for _, a := range elem.Attr {
					switch a.Name.Local {
					case "name":
						n = a.Value
					case "path":
						p = a.Value
					}
				}
				if n != "" && p != "" {
					ap := path.Join(root, p)
					project := RepoProject{name: n, path: p, abspath: ap}
					projects = append(projects, project)
				}
			}
		}
	}
	return
}

// MakeCmdString to create a string.
func MakeCmdString(tokens []string) (result string) {
	var buf bytes.Buffer
	for i, token := range tokens {
		if i > 0 {
			buf.WriteString(" ")
		}
		tokstr := QuoteCmdArg(token, "\"")
		buf.WriteString(tokstr)
	}
	result = buf.String()
	return
}

// QuoteCmdArg if it contains spaces or quotes.
//   s  - the string to quote
//   qc - the outer quote character to use (usually " or ').
func QuoteCmdArg(s string, qc string) string {
	qf := false
	ns := ""
	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch ch {
		case ' ', '\t':
			qf = true
		case '"', '\'':
			ns += "\\"
		}
		ns = ns + string(ch)
	}
	if qf {
		ns = qc + ns + qc
	}
	return ns
}

// GetExePath gets the path to an executable.
// exec.LookPath is not sufficient because it does not correctly interpret ~.
func GetExePath(f string) (a string, e error) {
	ax, ex := exec.LookPath(f)
	if ex == nil {
		a, _ = filepath.Abs(ax)
		return
	}

	// LookPath failed, get the value for ~ and look in $PATH.
	// This will not work on windows.

	switch runtime.GOOS {
	case "darwin", "linux":
		h := os.Getenv("HOME")
		if h == "" {
			e = fmt.Errorf("can't find path to %v, HOME environment variable not defined", f)
			return
		}

		p := os.Getenv("PATH")
		if p == "" {
			e = fmt.Errorf("can't find path to %v, PATH environment variable not defined", f)
			return
		}

		parts := strings.Split(p, ":")
		for _, part := range parts {
			pp := strings.Replace(part, "~", h, -1)
			fp := path.Join(pp, f)
			if _, err := os.Stat(fp); os.IsNotExist(err) {
				continue
			}
			a = fp
			return
		}
		e = fmt.Errorf("can't find %v in path: %v", f, p)
	default:
		e = fmt.Errorf("unsupported platform %v", runtime.GOOS)
		return
	}

	return
}

// TokenizeString - Dumb tokenizer for shell commands.
// It only recognizes white space and quotes.
func TokenizeString(text string) (tokens []string) {
	// Use two character modes to make it impossible to match
	// a single character.
	const WS = "ws"     // any whitespace
	const ACCEPT = "ac" // anything after a backslash
	const ESC = "\\"    // my editor complains about common strings

	// mode can be one of WS, ACCEPT, single-quote or double-quote.
	mode := WS

	// Tokens are captured in an array of strings.
	token := ""
	pch := "" // previous character
	pm := ""  // previous mode
	for i, c := range text {
		ch := string(text[i])
		switch mode {
		case WS:
			if unicode.IsSpace(c) {
				if len(token) > 0 {
					tokens = append(tokens, token)
					token = ""
				}
			} else {
				// not a space, check for a quote.
				switch ch {
				case "'":
					mode = ch
				case "\"":
					mode = ch
				case ESC:
					token += ESC
					pm = mode
					mode = ACCEPT
				default:
					token += ch
				}
			}
		case ch:
			// symmetric delimiter like " or '.
			if pch == ESC {
				token += ch
			} else {
				t := ch + token + ch
				tokens = append(tokens, t)
				token = ""
				mode = WS
			}
		case ACCEPT:
			// Accept anything after an escape.
			token += ch
			mode = pm
		default:
			// Everything else.
			token += ch
			if ch == ESC {
				// Handle the special case of escape sequences.
				pm = mode
				mode = ACCEPT
			}
		}
		pch = ch
	}
	if len(token) > 0 {
		tokens = append(tokens, token)
	}

	// Final pass to fix things up.
	for i, arg := range tokens {
		if arg[0] == '"' {
			tokens[i], _ = strconv.Unquote(arg)
		}
	}

	return
}
