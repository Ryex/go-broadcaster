package logutils

import (
	"os"

	logging "github.com/op/go-logging"
)

var Log *logging.Logger

var logFormat = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

func SetupLogging(name string, debug bool) {
	Log = logging.MustGetLogger("go-boradcaster")
	logStdout := logging.NewLogBackend(os.Stdout, "", 0)

	logStdoutFormat := logging.NewBackendFormatter(logStdout, logFormat)

	var logLevel logging.Level
	if logLevel = logging.INFO; debug {
		logLevel = logging.ERROR
	}

	logStdoutLeveled := logging.AddModuleLevel(logStdoutFormat)
	logStdoutLeveled.SetLevel(logLevel, "")

	logging.SetBackend(logStdoutLeveled)
}
