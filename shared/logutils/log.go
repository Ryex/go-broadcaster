package logutils

import (
	"os"

	logging "github.com/op/go-logging"
)

var Log = logging.MustGetLogger("example")

var logFormat = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

func SetupLogging() {
	logStdout := logging.NewLogBackend(os.Stdout, "", 0)
	logStderr := logging.NewLogBackend(os.Stderr, "", 0)

	logStderrFormat := logging.NewBackendFormatter(logStderr, logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	))

	logStderrLeveled := logging.AddModuleLevel(logStderrFormat)
	logStderrLeveled.SetLevel(logging.ERROR, "")

	logging.SetBackend(logStderrLeveled, logStdout)
}
