package main

import (
	"bytes"
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

// RepoProject contains the information for a repo project.
type RepoProject struct {
	name    string
	path    string
	abspath string
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
			a, _ = filepath.Abs(fp)
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
