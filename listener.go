package routemaster

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/deliveroo/routemaster-client-go/pkg/logmsg"
)

// The HandlerFunc type represents a function signature for consuming events
// received from Routemaster.
type HandlerFunc func([]*ReceivedEvent)

// A Listener is an implementation of http.Handler that handles Routemaster
// events.
type Listener struct {
	uuid    string
	handler HandlerFunc
	logger  logger
}

// NewListener creates a new handler for receiving Routemaster events.
func NewListener(uuid string, handler HandlerFunc, logger logger) *Listener {
	return &Listener{
		uuid:    uuid,
		handler: handler,
		logger:  defaultLogger(logger),
	}
}

// Returns a default logger that logs to stderr if nil.
func defaultLogger(l logger) logger {
	if l == nil {
		return log.New(os.Stderr, "routemaster: ", log.LstdFlags)
	}
	return l
}

// error replies to the request with the specified HTTP code, and a simple
// message based on the code.
func (l *Listener) error(w http.ResponseWriter, code int) {
	text := fmt.Sprintf("%d %s", code, http.StatusText(code))
	http.Error(w, text, code)
}

// handleEvents runs the handler, catching any panics and returning a
// 500.
func (l *Listener) handleEvents(w http.ResponseWriter, events []*ReceivedEvent) {
	defer func() {
		if err := recover(); err != nil {
			l.logger.Print(logmsg.Error("routemaster: panic during event handler").Set("error", err))
			l.error(w, http.StatusInternalServerError)
		}
	}()
	l.handler(events)
}

func (l *Listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username, _, _ := r.BasicAuth()
	if username != l.uuid {
		l.error(w, http.StatusUnauthorized)
		l.logger.Print(logmsg.Error("routemaster: unauthorized"))
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		l.error(w, http.StatusInternalServerError)
		l.logger.Print(logmsg.Error("routemaster: error reading body").SetError(err))
		return
	}

	var events []*ReceivedEvent
	err = json.Unmarshal(b, &events)
	if err != nil || len(events) == 0 {
		l.logger.Print(logmsg.Error("routemaster: bad request").Set("event", string(b)))
		l.error(w, http.StatusBadRequest)
		return
	}
	l.handleEvents(w, events)
}

// Allows clients of this library to pass a logger that satisfies this
// basic capability.
type logger interface {
	Print(v ...interface{})
}
