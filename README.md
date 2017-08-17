# Routemaster Client

`package routemaster`

A go API client and listener for the Routemaster event bus.

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

	c.Push(&routemaster.Event{
		Topic: "widgets",
		Type:  "create",
		URL:   "https://app.example.com/widgets/1",
		Data:  map[string]interface{}{
			"color": "teal",
		},
	})

#### Listen

To listen to events published on the bus:

	http.Handle("/", listener.New(func(events []*listener.Event) {
		fmt.Println(events[0].Topic)
	}))
	http.ListenAndServeTLS(":8123", "server.crt", "server.key", nil)
