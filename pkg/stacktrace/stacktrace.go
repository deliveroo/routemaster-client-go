// Package stacktrace provides simple and convenient wrappers for
// runtime.Callers and runtime.Stack.
package stacktrace

import (
	"runtime"
	"strings"
)

// Frame represents a function call in the call stack.
type Frame struct {
	Func string
	File string
	Line int
}

// Frames returns a list of stack frames for the current goroutine,
// skipping itself plus skipFrames frames above it. Returns at most
// maxFrames of the lowest frames.
func Frames(skipFrames, maxFrames int) []Frame {
	callers := make([]uintptr, maxFrames)
	n := runtime.Callers(skipFrames+2, callers)
	frames := runtime.CallersFrames(callers[:n])
	result := make([]Frame, 0, n)
	for {
		frame, more := frames.Next()
		if !more {
			break
		}
		if strings.HasPrefix(frame.Function, "runtime.") && frame.Function != "runtime.panic" {
			continue // omit (most) calls in runtime package
		}
		result = append(result, Frame{
			Func: frame.Function,
			File: frame.File,
			Line: frame.Line,
		})
	}
	return result
}

// Trace returns up to 4kb of the current goroutine's call stack.
func Trace() []byte {
	trace := make([]byte, 4096)
	n := runtime.Stack(trace, false)
	return trace[:n]
}
