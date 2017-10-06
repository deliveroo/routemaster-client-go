package logmsg

import (
	"bytes"
	"encoding/json"
	"runtime"
	"strconv"
	"time"

	"github.com/deliveroo/routemaster-client-go/pkg/glog"
	"github.com/deliveroo/routemaster-client-go/pkg/stacktrace"
)

// Message is a structured log message.
type Message struct {
	time    time.Time
	level   level
	file    string
	line    int
	trace   []stacktrace.Frame
	what    string
	context params
	data    params
}

type params map[string]interface{}

// Set adds a key-value pair to m.
func (m *Message) Set(k string, v interface{}) *Message {
	if m.data == nil {
		m.data = make(params, 1)
	}
	// Special case: if v has type error and does NOT implement the
	// json.Marshaler interface, store its Error() string. The built-in
	// error package returns values that marshal as "{}".
	if err, isError := v.(error); isError {
		if _, isJSONMarshaler := v.(json.Marshaler); !isJSONMarshaler {
			m.data[k] = err.Error()
			return m
		}
	}
	m.data[k] = v
	return m
}

// SetError calls Set("error", err).
func (m *Message) SetError(err error) *Message {
	m.Set("error", err)
	return m
}

// StackTrace adds the current goroutine's stack trace to m.
func (m *Message) StackTrace() *Message {
	m.trace = stacktrace.Frames(1, 10)
	return m
}

// String converts m to JSON format.
//
// Message also implements json.Marshaler, but this function constructs
// the JSON string manually for the following reasons: to render its
// fields in a consistent order; to precisely control the format of the
// Time field; so that if any user-defined parameter fails to marshal,
// it will not interfere with the remainder of the message.
//
func (m *Message) String() string {
	var b bytes.Buffer
	b.WriteString(`{"Time":"`)
	b.WriteString(m.time.Format(time.RFC3339Nano))
	b.WriteString(`","Level":"`)
	b.WriteString(m.level.String())
	b.WriteString(`","File":`)
	b.WriteString(strconv.Quote(m.file))
	b.WriteString(`,"Line":`)
	b.WriteString(strconv.Itoa(m.line))
	if len(m.trace) > 0 {
		b.WriteString(`,"Trace":`)
		data, _ := json.Marshal(m.trace)
		b.Write(data)
	}
	b.WriteString(`,"What":`)
	b.WriteString(strconv.Quote(m.what))
	if m.context != nil {
		b.WriteString(`,"Context":`)
		writeParams(&b, m.context)
	}
	if m.data != nil {
		b.WriteString(`,"Data":`)
		writeParams(&b, m.data)
	}
	b.WriteByte('}')
	return b.String()
}

// Marshals p as JSON, then writes it to b.
func writeParams(b *bytes.Buffer, p params) {
	first := true
	b.WriteByte('{')
	for k, v := range p {
		if !first {
			b.WriteByte(',')
		}
		first = false
		b.WriteString(strconv.Quote(k))
		b.WriteByte(':')
		bytes, err := json.Marshal(v)
		if err != nil {
			b.WriteString(strconv.Quote(err.Error()))
		} else {
			b.Write(bytes)
		}
	}
	b.WriteByte('}')
}

// MarshalJSON satisfies the json.Marshaler interface.
func (m *Message) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Time    time.Time
		Level   string             `json:",omitempty"`
		File    string             `json:",omitempty"`
		Line    int                `json:",omitempty"`
		Trace   []stacktrace.Frame `json:",omitempty"`
		What    string             `json:",omitempty"`
		Context params             `json:",omitempty"`
		Data    params             `json:",omitempty"`
	}{
		Time:    m.time,
		Level:   m.level.String(),
		File:    m.file,
		Line:    m.line,
		Trace:   m.trace,
		What:    m.what,
		Context: m.context,
		Data:    m.data,
	})
}

// Print calls glog.Print(m) with glog.Flags set to 0.
func (m *Message) Print() {
	flags := glog.Flags()
	glog.SetFlags(0)
	defer glog.SetFlags(flags)
	glog.Print(m)
}

// Fatal calls glog.Fatal(m) with glog.Flags set to 0.
func (m *Message) Fatal() {
	flags := glog.Flags()
	glog.SetFlags(0)
	defer glog.SetFlags(flags)
	glog.Fatal(m)
}

// Panic calls glog.Panic(m) with glog.Flags set to 0.
func (m *Message) Panic() {
	flags := glog.Flags()
	glog.SetFlags(0)
	defer glog.SetFlags(flags)
	glog.Panic(m)
}

func newMessage(level level, what string, ctx Context) *Message {
	_, file, line, _ := runtime.Caller(2)
	short := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
		file = short
	}
	return &Message{
		time:    time.Now(),
		level:   level,
		file:    short,
		line:    line,
		what:    what,
		context: params(ctx),
	}
}

func Debug(what string) *Message   { return newMessage(levelDebug, what, nil) }
func Info(what string) *Message    { return newMessage(levelInfo, what, nil) }
func Warning(what string) *Message { return newMessage(levelWarning, what, nil) }
func Error(what string) *Message   { return newMessage(levelError, what, nil) }

type Context params

// NewContext returns a new context.
func NewContext() Context {
	return Context(make(params, 1))
}

// Copy returns a deep copy of ctx.
func (ctx Context) Copy() Context {
	copy := make(params, len(ctx))
	for k, v := range ctx {
		copy[k] = v
	}
	return Context(copy)
}

// Set adds a key-value pair to ctx.
func (ctx Context) Set(k string, v interface{}) Context {
	ctx[k] = v
	return ctx
}

// Unset removes a key-value pair from ctx.
func (ctx Context) Unset(k string) Context {
	delete(ctx, k)
	return ctx
}

func (ctx Context) Info(what string) *Message    { return newMessage(levelInfo, what, ctx) }
func (ctx Context) Debug(what string) *Message   { return newMessage(levelDebug, what, ctx) }
func (ctx Context) Warning(what string) *Message { return newMessage(levelWarning, what, ctx) }
func (ctx Context) Error(what string) *Message   { return newMessage(levelError, what, ctx) }

type level int

const (
	levelDebug level = iota
	levelInfo
	levelWarning
	levelError
)

func (level level) String() string {
	switch level {
	case levelDebug:
		return "debug"
	case levelInfo:
		return "info"
	case levelWarning:
		return "warning"
	case levelError:
		return "error"
	default:
		return "???"
	}
}
