// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package log

import (
	"fmt"
	"log"
	"os"
)

/*
    Extended logging functionality providing Info, Warn and Error messages for
	log entries.
	Functions Infof(), Warnf(), Errorf() handles parameters in the same
	way that fmt.Printf handles parameters.
	Info(), Warn(), Error() take individual variables as parameters - useful
	for	printing of simple error message constants
*/

var infoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
var warnLogger = log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
var errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

// Info log entries print a message only, use for standard message output
// standard INFO print
func Info(v ...interface{}) {
	infoLogger.Output(2, fmt.Sprint(v...))
}

// INFO print with formatting
func Infof(format string, v ...interface{}) {
	infoLogger.Output(2, fmt.Sprintf(format, v...))
}

// standard INFO print with new line
func Infoln(v ...interface{}) {
	infoLogger.Output(2, fmt.Sprintln(v...))
}

// Warn log entries print a message only but wiht WARN marker, use for non-fatal
// errors
// standard WARN print
func Warn(v ...interface{}) {
	warnLogger.Output(2, fmt.Sprint(v...))
}

// WARN print with formatting
func Warnf(format string, v ...interface{}) {
	warnLogger.Output(2, fmt.Sprintf(format, v...))
}

// standard WARN print with new line
func Warnln(v ...interface{}) {
	warnLogger.Output(2, fmt.Sprintln(v...))
}

// Error log entries print a message then halt execution
// standard ERROR print.
func Error(v ...interface{}) {
	errorLogger.Output(2, fmt.Sprint(v...))
	os.Exit(1)
}

// ERROR print with formatting
func Errorf(format string, v ...interface{}) {
	errorLogger.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}
