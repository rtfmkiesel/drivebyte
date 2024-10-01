package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

var (
	ShowDebugOutput = false // If set to true, logger.Debug() messages will get printed
	styleDebug      = "[" + color.BlueString("DBG") + "]"
	styleWarning    = "[" + color.YellowString("WAR") + "]"
	styleError      = "[" + color.RedString("ERR") + "]"
	styleFatal      = "[" + color.RedString("FTL") + "]"
)

// Stdout print a format string to stdout
func Stdout(s string, args ...interface{}) {
	msg := fmt.Sprintf(s, args...)
	fmt.Fprint(color.Output, formatMsg(msg))
}

// Debug prints a debug/info format string to stderr
func Debug(s string, args ...interface{}) {
	if !ShowDebugOutput {
		return
	}

	msg := fmt.Sprintf(s, args...)
	msg = fmt.Sprintf("%s %s", styleDebug, msg)

	fmt.Fprint(color.Error, formatMsg(msg))
}

// Warning prints a warning format string to stderr
func Warning(s string, args ...interface{}) {
	msg := fmt.Sprintf(s, args...)
	msg = fmt.Sprintf("%s %s", styleWarning, msg)

	fmt.Fprint(color.Error, formatMsg(msg))
}

// Error prints an error to stderr
func Error(err error) {
	msg := fmt.Sprintf("%s %s", styleError, err)
	fmt.Fprint(color.Error, formatMsg(msg))
}

// Fatal prints an fatal error to stderr and quits
func Fatal(err error) {
	msg := fmt.Sprintf("%s %s", styleFatal, err)
	fmt.Fprint(color.Error, formatMsg(msg))
	os.Exit(1)
}

func formatMsg(s string) string {
	if strings.HasSuffix(s, "\n") {
		return s
	} else {
		return s + "\n"
	}
}
