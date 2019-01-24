package logutils

import (
	"io"
	"os"

	logging "github.com/op/go-logging"
)

// Log is the logging global singelton
var Log *logging.Logger

var logFormat = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

// SetupLogging sets up logging backend so that logutils can be useed
// - name - the name to give to the logger SetBackend
// - debug - shoudl the loggin be done in debug model
// - out - the output backend for the logger
func SetupLogging(name string, debug bool, out io.Writer) {

	if out == nil {
		out = os.Stdout
	}

	Log = logging.MustGetLogger(name)
	logStdout := logging.NewLogBackend(out, "", 0)

	logStdoutFormat := logging.NewBackendFormatter(logStdout, logFormat)

	logLevel := logging.INFO
	if !debug {
		logLevel = logging.ERROR
	}

	logStdoutLeveled := logging.AddModuleLevel(logStdoutFormat)
	logStdoutLeveled.SetLevel(logLevel, "")

	logging.SetBackend(logStdoutLeveled)
}
