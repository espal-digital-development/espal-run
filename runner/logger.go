package runner

import (
	"fmt"
	"os"
	"time"

	"github.com/juju/errors"
)

type logFunc func(string, ...interface{})

func (r *Runner) newLogFunc(prefix string) func(string, ...interface{}) {
	color, clear := "", ""
	if r.config.Colorize {
		color = fmt.Sprintf("\033[%sm", r.logColor(prefix))
		clear = fmt.Sprintf("\033[%sm", r.logColors["reset"])
	}
	prefix = fmt.Sprintf("%-11s", prefix)

	return func(format string, v ...interface{}) {
		now := time.Now()
		timeString := fmt.Sprintf("%d:%d:%02d", now.Hour(), now.Minute(), now.Second())
		format = fmt.Sprintf("%s%s %s |%s %s", color, timeString, prefix, clear, format)
		r.logger.Printf(format, v...)
	}
}

func (r *Runner) fatal(err error) {
	r.logger.Fatal(errors.ErrorStack(err))
}

type appLogWriter struct {
	r *Runner
}

func (a appLogWriter) Write(p []byte) (n int, err error) {
	a.r.appLog(string(p))

	return len(p), nil
}

func (r *Runner) logColor(logName string) string {
	switch logName {
	case "main":
		return r.logColors[r.config.LogColors.Main]
	case "watcher":
		return r.logColors[r.config.LogColors.Watcher]
	case "runner":
		return r.logColors[r.config.LogColors.Runner]
	case "build":
		return r.logColors[r.config.LogColors.Build]
	case "app":
		return r.logColors[r.config.LogColors.App]
	default:
		r.fatal(errors.Errorf("unsupported logName `%s`", logName))
	}
	return ""
}

func (r *Runner) createBuildErrorsLog(message string) bool {
	file, err := os.Create(r.buildErrorsFilePath())
	if err != nil {
		return false
	}
	if _, err := file.WriteString(message); err != nil {
		return false
	}
	return true
}

func (r *Runner) removeBuildErrorsLog() error {
	return errors.Trace(os.Remove(r.buildErrorsFilePath()))
}
