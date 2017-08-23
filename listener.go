package routemaster

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// The HandlerFunc type represents a function signature for consuming events
// received from Routemaster.
type HandlerFunc func([]*ReceivedEvent)

// A Listener is an implementation of http.Handler that handles Routemaster
// events.
type Listener struct {
	uuid    string
	handler HandlerFunc
}

// NewListener creates a new handler for receiving Routemaster events.
func NewListener(uuid string, handler HandlerFunc) *Listener {
	return &Listener{
		uuid:    uuid,
		handler: handler,
	}
}

// error replies to the request with the specified HTTP code, and a simple
// message based on the code.
func (l *Listener) error(w http.ResponseWriter, code int) {
	text := fmt.Sprintf("%d %s", code, http.StatusText(code))
	http.Error(w, text, code)
}

// handleEvents runs the handler, catching any panics and returning a 500.
func (l *Listener) handleEvents(w http.ResponseWriter, events []*ReceivedEvent) {
	defer func() {
		if err := recover(); err != nil {
			l.error(w, http.StatusInternalServerError)
		}
	}()
	l.handler(events)
}

func (l *Listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username, _, _ := r.BasicAuth()
	if username != l.uuid {
		l.error(w, http.StatusUnauthorized)
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		l.error(w, http.StatusInternalServerError)
		return
	}

	var events []*ReceivedEvent
	err = json.Unmarshal(b, &events)
	if err != nil || len(events) == 0 {
		l.error(w, http.StatusBadRequest)
		return
	}
	l.handleEvents(w, events)
}
