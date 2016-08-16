package main

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/jlinoff/go/run"
)

// ContextInfoStruct contains context information for the run.
type ContextInfoStruct struct {
	Base          string
	Cmd           string
	Dir           string
	ExePath       string
	HostName      string
	MakeBuildDate string
	MakeVersion   string
	NumCpus       int
	OsType        string
	OsVersion     string
	RecipeDir     string
	ScriptDir     string
	Tee           string
	Time          string
	TimeStamp     string
	UserGID       int
	UserName      string
	UserPID       int
	UserUID       int
}

// Context is the context object.
var Context ContextInfoStruct

// MakeContext makes the context object with the context information.
func MakeContext(tee string) {
	u, _ := user.Current()
	d, _ := os.Getwd()
	h, _ := os.Hostname()
	t := time.Now().Truncate(time.Millisecond).Format("2006-01-02 15:05:05")
	ts := time.Now().Truncate(time.Millisecond).Format("20060102-150505")
	ep, err := GetExePath(os.Args[0])
	if err != nil {
		Log.Err("internal error - can't find path for %v - %v", os.Args[0], err)
	}
	b := filepath.Base(ep)
	rp := path.Join(filepath.Dir(ep), "../etc", b, "recipes")
	base := filepath.Base(os.Args[0])

	osver := "?"
	switch runtime.GOOS {
	case "linux", "darwin":
		cout, _ := run.CmdSilent(TokenizeString("uname -m -r -s -v"))
		osver = strings.TrimSpace(cout)
	}

	// Define the directory where anonymous scripts are stored.
	// We can't use /tmp because we can't guarantee that execution mode is enabled.
	// We could simply force the user to use bash (or whatever) and execute the
	// the scripts as text files but this limits the type of scripts which are
	// supported.
	// The approach of using the users $HOME directory is the most flexible.
	sd := fmt.Sprintf("%v/.%v", os.Getenv("HOME"), base)

	Context.Base = base
	Context.Cmd = MakeCmdString(os.Args)
	Context.Dir = d
	Context.ExePath = ep
	Context.HostName = h
	Context.NumCpus = runtime.NumCPU()
	Context.OsType = runtime.GOOS
	Context.OsVersion = osver
	Context.RecipeDir = rp
	Context.ScriptDir = sd
	Context.Tee = tee
	Context.Time = t
	Context.TimeStamp = ts
	Context.UserGID = os.Getgid()
	Context.UserName = u.Username
	Context.UserPID = os.Getpid()
	Context.UserUID = os.Getuid()
	Context.MakeVersion = Version     // generated by make
	Context.MakeBuildDate = BuildDate // generated by make

	return
}

// PrintContext prints the container information.
func (info ContextInfoStruct) PrintContext() {
	Log.Info("context")
	Log.Info("   base     : %v", info.Base)
	Log.Info("   cmd      : %v", info.Cmd)
	Log.Info("   directory: %v", info.Dir)
	Log.Info("   exe      : %v", info.ExePath)
	Log.Info("   gid      : %v", info.UserGID)
	Log.Info("   numcpus  : %v", info.NumCpus)
	Log.Info("   os       : %v", info.OsType)
	Log.Info("   osver    : %v", info.OsVersion)
	Log.Info("   pid      : %v", info.UserPID)
	Log.Info("   recipes  : %v", info.RecipeDir)
	Log.Info("   scripts  : %v", info.ScriptDir)
	if info.Tee != "" {
		Log.Info("   tee      : %v", info.Tee)
	}
	Log.Info("   time     : %v", info.Time)
	Log.Info("   timestamp: %v", info.TimeStamp)
	Log.Info("   uid      : %v", info.UserUID)
	Log.Info("   user     : %v", info.UserName)
	Log.Info("   version  : %v - %v", info.MakeVersion, info.MakeBuildDate)
}
