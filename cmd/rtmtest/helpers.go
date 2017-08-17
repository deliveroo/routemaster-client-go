package main

import (
	"net/http"
	"time"

	routemaster "github.com/deliveroo/routemaster-client-go"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func newClient(token string) *routemaster.Client {
	c, err := routemaster.NewClient(&routemaster.Config{
		URL:  config.url,
		UUID: token,
	})
	must(err)
	return c
}

type listener struct {
	ch chan []*routemaster.Event
}

func startListener(addr, crt, key, auth string) *listener {
	l := &listener{
		ch: make(chan []*routemaster.Event, 1),
	}
	rml := routemaster.NewListener(auth, func(events []*routemaster.Event) {
		l.ch <- events
	})
	go func() {
		must(http.ListenAndServeTLS(addr, crt, key, rml))
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
