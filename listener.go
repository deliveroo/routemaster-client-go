package routemaster

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// The HandlerFunc type represents a function signature for consuming events
// received from Routemaster.
type HandlerFunc func([]*Event)

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

func (l *Listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username, _, _ := r.BasicAuth()
	if username != l.uuid {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var events []*Event
	err = json.Unmarshal(b, &events)
	if err != nil || len(events) == 0 {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	l.handler(events)
}
