package stacktrace_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/deliveroo/routemaster-client-go/pkg/stacktrace"
)

func catch() (frames []stacktrace.Frame) {
	defer func() {
		if r := recover(); r != nil {
			frames = stacktrace.Frames(1, 99)
		}
	}()
	f0(2)
	return
}

func f0(n int) {
	if n%2 == 0 {
		f1(n + 1)
	}
}

func f1(n int) {
	if n%2 == 1 {
		f2(n + 1)
	}
}

func f2(n int) {
	if n%2 == 0 {
		panic("n is even")
	}
}

func TestStacktracePanic(t *testing.T) {
	frames := catch()
	for i := range frames {
		frames[i].Func = lastPart(frames[i].Func)
		frames[i].File = lastPart(frames[i].File)
	}
	expected := []stacktrace.Frame{
		{
			Func: "stacktrace_test.f2",
			File: "stacktrace_test.go",
			Line: 35,
		},
		{
			Func: "stacktrace_test.f1",
			File: "stacktrace_test.go",
			Line: 29,
		},
		{
			Func: "stacktrace_test.f0",
			File: "stacktrace_test.go",
			Line: 23,
		},
		{
			Func: "stacktrace_test.catch",
			File: "stacktrace_test.go",
			Line: 17,
		},
		{
			Func: "stacktrace_test.TestStacktracePanic",
			File: "stacktrace_test.go",
			Line: 40,
		},
		{
			Func: "testing.tRunner",
			File: "testing.go",
			Line: 746,
		},
	}
	if !reflect.DeepEqual(frames, expected) {
		t.Fatalf("stacktrace is %+v, want %+v", frames, expected)
	}
}

func lastPart(s string) string {
	i := strings.LastIndex(s, "/")
	if i < 0 {
		return s
	}
	return s[i+1:]
}
