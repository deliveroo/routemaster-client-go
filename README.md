# Routemaster Client

[![Build Status](https://travis-ci.com/deliveroo/routemaster-client-go.svg?token=Cn6Bjq9ZZa5MrmKhd9RW&branch=master)](https://travis-ci.com/deliveroo/routemaster-client-go)

`package routemaster`

A go API client and listener for the Routemaster event bus.

**Note:** This is a first draft, and the API is subject to change.

## Install

Since this is (currently) a private repo, you first need to configure `go get`
to fetch via SSH rather than HTTPS:

	git config --global url."git@github.com:".insteadOf "https://github.com/"

Then you can install via `go get` as usual:

	go get github.com/deliveroo/routemaster-client-go

## Usage

First, create a new client:

	c, err := routemaster.NewClient(&routemaster.Config{
		URL:  "https://routemaster.dev",
		UUID: "demo",
	})

### Commands

#### Subscribe

To subscribe to one or more topics:

	c.Subscribe(&routemaster.Subscription{
		Topics:   []string{"widgets"},
		Callback: "https://app.example.com/events",
		UUID:     "demo",
	}))

#### Unsubscribe

To unsubscribe from a single topic:

	c.Unsubscribe("widgets")

To unsubscribe from all topics:

	c.UnsubscribeAll()

#### Push

To push an event to the bus:

	c.Push("widgets", &routemaster.Event{
		Type:  "create",
		URL:   "https://app.example.com/widgets/1",
		Data:  map[string]interface{}{
			"color": "teal",
		},
	})

#### Listen

To listen to events published on the bus:

	http.Handle("/", routemaster.NewListener(
		"demo",
		func(events []*routemaster.ReceivedEvent) {
			for _, e := range events {
				log.Printf("%v\n", e)
			}
		})
	))
	http.ListenAndServeTLS(":8123", "server.crt", "server.key", nil)
