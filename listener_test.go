package routemaster

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListener(t *testing.T) {
	t.Run("successful event", func(t *testing.T) {
		r := newTestRunner("secret")
		r.do("/events", "secret", `
			[{
				"topic": "orders",
				"type": "create",
				"url": "https://orders/1",
				"t": 500
			}]
		`)

		if want := http.StatusOK; r.response.StatusCode != want {
			t.Errorf("status: got %d, want %d", r.response.StatusCode, want)
			t.FailNow()
		}

		if len(r.events) == 0 {
			t.Error("no event received")
			t.FailNow()
		}

		if want := "orders"; r.events[0].Topic != want {
			t.Errorf("topic: got %q, want %q", r.events[0].Topic, want)
		}
		if want := "create"; r.events[0].Type != want {
			t.Errorf("type: got %q, want %q", r.events[0].Type, want)
		}
		if want := "https://orders/1"; r.events[0].URL != want {
			t.Errorf("url: got %q, want %q", r.events[0].URL, want)
		}
		if want := int64(500); r.events[0].Timestamp != want {
			t.Errorf("t: got %q, want %q", r.events[0].Timestamp, want)
		}
		if want, got := "", r.readBody(); want != got {
			t.Errorf("body: got %q, want %q", got, want)
		}
	})

	t.Run("bad event json", func(t *testing.T) {
		r := newTestRunner("secret")
		r.do("/events", "secret", `
			{
				"topic": "orders",
				"type": "create",
				"url": "https://orders/1"
			}]
		`)

		if want := http.StatusBadRequest; r.response.StatusCode != want {
			t.Errorf("status: got %d, want %d", r.response.StatusCode, want)
			t.FailNow()
		}
		if want, got := "400 Bad Request\n", r.readBody(); want != got {
			t.Errorf("body: got %q, want %q", got, want)
		}
	})

	t.Run("bad auth", func(t *testing.T) {
		r := newTestRunner("secret")
		r.do("/events", "wrong password", "")

		if want := http.StatusUnauthorized; r.response.StatusCode != want {
			t.Errorf("status: got %d, want %d", r.response.StatusCode, want)
			t.FailNow()
		}
		if want, got := "401 Unauthorized\n", r.readBody(); want != got {
			t.Errorf("body: got %q, want %q", got, want)
		}
	})

	t.Run("panic in handler", func(t *testing.T) {
		r := newTestRunner("secret")
		r.panic = true
		r.do("/events", "secret", `
			[{
				"topic": "orders",
				"type": "create",
				"url": "https://orders/1"
			}]
		`)
		if want := http.StatusInternalServerError; r.response.StatusCode != want {
			t.Errorf("status: got %d, want %d", r.response.StatusCode, want)
			t.FailNow()

		}
		if want, got := "500 Internal Server Error\n", r.readBody(); want != got {
			t.Errorf("body: got %q, want %q", got, want)
		}
	})
}

type testRunner struct {
	uuid     string
	panic    bool
	events   []*ReceivedEvent
	response *http.Response
}

func newTestRunner(uuid string) *testRunner {
	return &testRunner{uuid: uuid}
}

func (t *testRunner) do(url, username, body string) {
	listener := NewListener(t.uuid, func(events []*ReceivedEvent) {
		if t.panic {
			panic(errors.New("unknown error"))
		}
		t.events = events
	})
	ts := httptest.NewServer(listener)
	req, err := http.NewRequest(http.MethodPost, ts.URL+url, bytes.NewBufferString(body))
	req.SetBasicAuth(username, "")
	must(err)
	response, err := http.DefaultClient.Do(req)
	must(err)
	t.response = response
}

func (t *testRunner) readBody() string {
	body, err := ioutil.ReadAll(t.response.Body)
	must(err)
	return string(body)
}
