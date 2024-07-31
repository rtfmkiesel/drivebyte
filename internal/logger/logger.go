package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

var (
	DebugOutput  = false
	styleInfo    = "[" + color.BlueString("INF") + "]"
	styleWarning = "[" + color.YellowString("WAR") + "]"
	styleError   = "[" + color.RedString("ERR") + "]"
	styleFatal   = "[" + color.RedString("FTL") + "]"
)

// Stdout prints a format string message to stdout
func Stdout(msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	fmt.Fprint(color.Output, msg)
}

// Info prints a info format string to stderr
func Info(s string, args ...interface{}) {
	if !DebugOutput {
		return
	}

	msg := fmt.Sprintf(s, args...)
	msg = fmt.Sprintf("%s %s", styleInfo, msg)

	fmt.Fprint(color.Error, formatMsg(msg))
}

// Warning prints a warning format string to stderr
func Warning(s string, args ...interface{}) {
	if !DebugOutput {
		return
	}

	msg := fmt.Sprintf(s, args...)
	msg = fmt.Sprintf("%s %s", styleWarning, msg)

	fmt.Fprint(color.Error, formatMsg(msg))
}

// ErrorF prints an error format string to stderr
func ErrorF(s string, args ...interface{}) {
	msg := fmt.Sprintf(s, args...)
	msg = fmt.Sprintf("%s %s", styleError, msg)
	fmt.Fprint(color.Error, formatMsg(msg))
}

// Fatal prints an fatal error to stderr and quit
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
