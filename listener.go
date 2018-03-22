package routemaster

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// The HandlerFunc type represents a function signature for consuming events
// received from Routemaster.
type HandlerFunc func([]*ReceivedEvent) error

// The onError represents a function signature for handling
// errors emitted by the listener.
type onError func(error)

// ListenerConfig specifies the way a listener should be set up.
type ListenerConfig struct {
	Handler HandlerFunc
	Logger  *log.Logger
	OnError onError
	UUID    string
}

// A Listener is an implementation of http.Handler that handles Routemaster
// events.
type Listener struct {
	handler HandlerFunc
	logger  *log.Logger
	onError onError
	uuid    string
}

// NewListener creates a new handler for receiving Routemaster events.
func NewListener(cfg *ListenerConfig) *Listener {
	return &Listener{
		handler: cfg.Handler,
		logger:  defaultLogger(cfg.Logger),
		onError: cfg.OnError,
		uuid:    cfg.UUID,
	}
}

// reportError replies to the request with the specified HTTP code, and a simple
// message based on the code.
func (l *Listener) reportError(w http.ResponseWriter, code int, err error) {
	var wroteError bool

	// If onError panics, log the panic and continue to send the
	// intended status code back.
	defer func() {
		if r := recover(); r != nil {
			l.logger.Printf("panic while running error handler: %+v", r)
			if !wroteError {
				text := fmt.Sprintf("%d %s", code, http.StatusText(code))
				http.Error(w, text, code)
			}
		}
	}()

	if l.onError != nil {
		l.onError(err)
	}

	text := fmt.Sprintf("%d %s", code, http.StatusText(code))
	http.Error(w, text, code)
	wroteError = true
}

func (l *Listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("%+v", r)
			}
			l.reportError(w, http.StatusInternalServerError, err)
			return
		}
	}()

	// Check for the expected username.
	username, _, _ := r.BasicAuth()
	if username != l.uuid {
		l.reportError(w, http.StatusUnauthorized, errors.New("bad token"))
		return
	}

	// Read the request body.
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		l.reportError(w, http.StatusInternalServerError, errors.New("request body read failed"))
		return
	}

	// Unmarshal received events.
	var events []*ReceivedEvent
	err = json.Unmarshal(b, &events)
	if err != nil || len(events) == 0 {
		l.reportError(w, http.StatusBadRequest, fmt.Errorf("body malformed: %s", string(b)))
		return
	}

	// Finally, handle events.
	if err := l.handler(events); err != nil {
		l.reportError(w, http.StatusInternalServerError, err)
		return
	}
}

// defaultLogger returns a logger if the given one is nil.
func defaultLogger(logger *log.Logger) *log.Logger {
	if logger != nil {
		return logger
	}
	return log.New(os.Stderr, "routemaster/client: ", log.LstdFlags)
}
