// logger handles a super simple logger
package logger

import (
	"fmt"
	"os"
	"strings"
)

var (
	Verbose bool = false
)

func Info(msg string, args ...interface{}) {
	log("[*] "+msg, args...)
}

func Success(msg string, args ...interface{}) {
	log("[+] "+msg, args...)
}

func Error(msg string, args ...interface{}) {
	// Not via log() since errors should always be printed
	msg = fmt.Sprintf("[!] "+msg, args...)
	if strings.HasSuffix(msg, "\n") {
		fmt.Fprint(os.Stderr, msg)
	} else {
		fmt.Fprint(os.Stderr, msg+"\n")
	}
}

func log(msg string, args ...interface{}) {
	if !Verbose {
		return
	}

	if strings.HasSuffix(msg, "\n") {
		fmt.Fprintf(os.Stderr, msg, args...)
	} else {
		fmt.Fprintf(os.Stderr, msg+"\n", args...)
	}
}
