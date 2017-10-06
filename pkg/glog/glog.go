// Package glog provides a global log.Logger instance. It exists only
// because the stdlib log package does not expose its default logger.
package glog

import (
	"io"
	"log"
	"os"
)

// Logger is the logger that any subsystem should use if it was not
// given a specific logger to use instead. The standard logger should be
// avoided; that is, do not call log.Printf, log.Fatal, and so on.
var Logger = log.New(os.Stderr, "", log.LstdFlags)

// DefaultLogger returns Logger if l is nil. Otherwise it returns l. It
// is a convenient shorthand for
//
//	if l == nil {
//		return Logger
//	}
//	return l
//
func DefaultLogger(l *log.Logger) *log.Logger {
	if l == nil {
		return Logger
	}
	return l
}

// These functions mirror the log package's public interface and simply
// delegate to our default global logger:

func Fatal(v ...interface{})                 { Logger.Fatal(v...) }
func Fatalf(format string, v ...interface{}) { Logger.Fatalf(format, v...) }
func Fatalln(v ...interface{})               { Logger.Fatalln(v...) }
func Flags() int                             { return Logger.Flags() }
func Output(calldepth int, s string) error   { return Logger.Output(calldepth, s) }
func Panic(v ...interface{})                 { Logger.Panic(v...) }
func Panicf(format string, v ...interface{}) { Logger.Panicf(format, v...) }
func Panicln(v ...interface{})               { Logger.Panicln(v...) }
func Prefix() string                         { return Logger.Prefix() }
func Print(v ...interface{})                 { Logger.Print(v...) }
func Printf(format string, v ...interface{}) { Logger.Printf(format, v...) }
func Println(v ...interface{})               { Logger.Println(v...) }
func SetFlags(flag int)                      { Logger.SetFlags(flag) }
func SetOutput(w io.Writer)                  { Logger.SetOutput(w) }
func SetPrefix(prefix string)                { Logger.SetPrefix(prefix) }
