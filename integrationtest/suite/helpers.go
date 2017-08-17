package suite

import (
	"encoding/json"
	"net/http"
	"time"

	routemaster "github.com/deliveroo/routemaster-client-go"
)

// Panics if err is not nil.
func must(err error) {
	if err != nil {
		panic(err)
	}
}

// Compares the JSON representation of two values, to avoid false
// negatives (e.g. float vs. int, array vs. slice).
func eqJSON(a, b interface{}) bool {
	aBytes, _ := json.Marshal(a)
	bBytes, _ := json.Marshal(b)
	return string(aBytes) == string(bBytes)
}

type listener struct {
	ch     chan []*routemaster.Event
	server *http.Server
}

func startListener(addr, crt, key, auth string) *listener {
	l := &listener{
		ch: make(chan []*routemaster.Event, 1),
	}
	rml := routemaster.NewListener(auth, func(events []*routemaster.Event) {
		l.ch <- events
	})
	go func() {
		l.server = &http.Server{Addr: addr, Handler: rml}
		l.server.ListenAndServeTLS(crt, key)
	}()
	return l
}

func (l *listener) waitForEvent(timeout time.Duration) *routemaster.Event {
	select {
	case events := <-l.ch:
		return events[0]
	case <-time.After(timeout):
		return nil
	}
}

func (l *listener) close() {
	l.server.Close()
}
