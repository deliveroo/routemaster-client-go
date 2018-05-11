# Routemaster Client

[![Build Status](https://travis-ci.com/deliveroo/routemaster-client-go.svg?token=Cn6Bjq9ZZa5MrmKhd9RW&branch=master)](https://travis-ci.com/deliveroo/routemaster-client-go)

`package routemaster`

A Go API client and listener for the Routemaster event bus.

## Usage

First, create a new client:

```go
c, err := routemaster.NewClient(&routemaster.Config{
    URL:  "https://routemaster.dev",
    UUID: "demo",
})
```

### Commands

#### Subscribe

To subscribe to one or more topics:

```go
c.Subscribe(&routemaster.Subscription{
    Topics:   []string{"widgets"},
    Callback: "https://app.example.com/events",
    UUID:     "demo",
}))
```

#### Unsubscribe

To unsubscribe from a single topic:

```go
c.Unsubscribe("widgets")
```

To unsubscribe from all topics:

```go
c.UnsubscribeAll()
```

#### Push

To push an event to the bus:

```go
c.Push("widgets", &routemaster.Event{
    Type:  "create",
    URL:   "https://app.example.com/widgets/1",
    Data:  map[string]interface{}{
        "color": "teal",
    },
})
```

#### Listen

To listen to events published on the bus:

```go
http.Handle("/", routemaster.NewListener(
    "demo",
    func(events []*routemaster.ReceivedEvent) {
        for _, e := range events {
            log.Printf("%v\n", e)
        }
    })
))
http.ListenAndServeTLS(":8123", "server.crt", "server.key", nil)
```
