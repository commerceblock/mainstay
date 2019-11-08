// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package log

import (
    "log"
    "os"
    "fmt"
)

/* 	Extended logging functionality providing Info, Warn and Error messages for
	log entries.
	Functions Infof(), Warnf(), Errorf() handles parameters in the same
	way that fmt.Printf handles parameters.
	Info(), Warn(), Error() take individual variables as parameters - useful
	for	printing of simple error message constants
*/

var infoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
var warnLogger = log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
var errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)


func Info(v ...interface{}) {
    infoLogger.Output(2, fmt.Sprint(v...))
}
func Infof(format string, v ...interface{}) {
    infoLogger.Output(2, fmt.Sprintf(format, v...))
}
func Infoln(v ...interface{}) {
    infoLogger.Output(2, fmt.Sprintln(v...))
}

func Warn(v ...interface{}) {
    warnLogger.Output(2, fmt.Sprint(v...))
}
func Warnf(format string, v ...interface{}) {
    warnLogger.Output(2, fmt.Sprintf(format, v...))
}
func Warnln(v ...interface{}) {
    warnLogger.Output(2, fmt.Sprintln(v...))
}

func Error(v ...interface{}) {
	errorLogger.Output(2, fmt.Sprint(v...))
	os.Exit(1)
}
func Errorf(format string, v ...interface{}) {
    errorLogger.Output(2, fmt.Sprintf(format, v...))
    os.Exit(1)
}
func Errorln(v ...interface{}) {
    errorLogger.Output(2, fmt.Sprintln(v...))
}
