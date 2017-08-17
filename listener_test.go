package routemaster

import (
	"bytes"
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
				"url": "https://orders/1"
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
	})

	t.Run("bad auth", func(t *testing.T) {
		r := newTestRunner("secret")
		r.do("/events", "wrong password", "")

		if want := http.StatusUnauthorized; r.response.StatusCode != want {
			t.Errorf("status: got %d, want %d", r.response.StatusCode, want)
			t.FailNow()
		}
	})
}

type testRunner struct {
	uuid     string
	events   []*Event
	response *http.Response
}

func newTestRunner(uuid string) *testRunner {
	return &testRunner{uuid: uuid}
}

func (t *testRunner) do(url, username, body string) {
	listener := NewListener(t.uuid, func(events []*Event) {
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
