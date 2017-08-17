package suite

import (
	"reflect"
	"time"

	routemaster "github.com/deliveroo/routemaster-client-go"
	"github.com/deliveroo/routemaster-client-go/integrationtest"
)

var allTests = []*integrationtest.Test{
	{"testRoundTrip", testRoundTrip},
	{"testUnsubscribe", testUnsubscribe},
}

func testRoundTrip(t *integrationtest.T) {
	token, err := t.RootClient().CreateToken("client")
	must(err)

	client := t.NewClient(token)

	const listenerSecret = "secret"
	must(client.Subscribe(&routemaster.Subscription{
		Topics:   []string{"orders"},
		UUID:     listenerSecret,
		Callback: "https://localhost:8080",
	}))

	listener := startListener(":8080", "cmd/rtmtest/server.crt", "cmd/rtmtest/server.key", listenerSecret)
	defer listener.close()

	sent := &routemaster.Event{
		Topic: "orders",
		Type:  "create",
		URL:   "https://orderweb/orders/1",
	}
	must(client.Push(sent))

	received := listener.waitForEvent(1 * time.Second)
	if received == nil {
		t.Error("no event received")
		return
	}
	if !reflect.DeepEqual(received, sent) {
		t.Errorf("event: got %+v, want %+v", received, sent)
	}
}

func testUnsubscribe(t *integrationtest.T) {
	token, err := t.RootClient().CreateToken("client")
	must(err)

	client := t.NewClient(token)
	const listenerSecret = "secret"
	must(client.Subscribe(&routemaster.Subscription{
		Topics:   []string{"riders"},
		UUID:     listenerSecret,
		Callback: "https://localhost:8080",
	}))

	must(client.UnsubscribeAll())

	listener := startListener(":8080", "cmd/rtmtest/server.crt", "cmd/rtmtest/server.key", listenerSecret)
	defer listener.close()

	sent := &routemaster.Event{
		Topic: "riders",
		Type:  "create",
		URL:   "https://orderweb/riders/1",
	}
	must(client.Push(sent))

	received := listener.waitForEvent(1 * time.Second)
	if received != nil {
		t.Errorf("event received after unsubscribing")
	}
}
