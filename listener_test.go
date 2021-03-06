package routemaster

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
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
			t.Fatalf("status: got %d, want %d", r.response.StatusCode, want)
		}

		if len(r.events) == 0 {
			t.Fatal("no event received")
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

	t.Run("data payload", func(t *testing.T) {
		r := newTestRunner("secret")
		r.do("/events", "secret", `
			[{
				"topic": "orders",
				"type": "create",
				"url": "https://orders/1",
				"t": 500,
				"data": { "restaurant_id": 123 }
			}]
		`)

		if want := http.StatusOK; r.response.StatusCode != want {
			t.Fatalf("status: got %d, want %d", r.response.StatusCode, want)
		}

		if len(r.events) == 0 {
			t.Fatal("no event received")
		}

		event := r.events[0]
		var data struct {
			RestaurantID int `json:"restaurant_id"`
		}
		must(json.Unmarshal(event.Data, &data))

		if data.RestaurantID != 123 {
			t.Errorf("data.RestaurantID: got %d, want %d", data.RestaurantID, 123)
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
			t.Fatalf("status: got %d, want %d", r.response.StatusCode, want)
		}
		if want, got := "400 Bad Request\n", r.readBody(); want != got {
			t.Errorf("body: got %q, want %q", got, want)
		}

		// Logs should contain "bad request" and the JSON that was
		// passed in.
		logs := r.readLogs()
		if !strings.Contains(logs, "body malformed") || !strings.Contains(logs, "https://orders/1") {
			t.Errorf("logs: missing bad request or bad input, got: %s", logs)
		}
	})

	t.Run("bad auth", func(t *testing.T) {
		r := newTestRunner("secret")
		r.do("/events", "wrong password", "")

		if want := http.StatusUnauthorized; r.response.StatusCode != want {
			t.Fatalf("status: got %d, want %d", r.response.StatusCode, want)
		}
		if want, got := "401 Unauthorized\n", r.readBody(); want != got {
			t.Errorf("body: got %q, want %q", got, want)
		}

		wantLog := "bad token"
		if !strings.Contains(r.readLogs(), wantLog) {
			t.Errorf("logs: got %q, want %q", r.readLogs(), wantLog)
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
			t.Fatalf("status: got %d, want %d", r.response.StatusCode, want)

		}
		if want, got := "500 Internal Server Error\n", r.readBody(); want != got {
			t.Errorf("body: got %q, want %q", got, want)
		}
		wantLog := "panic error"
		if !strings.Contains(r.readLogs(), wantLog) {
			t.Errorf("logs: got %q, want %q", r.readLogs(), wantLog)
		}
	})

	t.Run("error in handler", func(t *testing.T) {
		r := newTestRunner("secret")
		r.error = true
		r.do("/events", "secret", `
			[{
				"topic": "orders",
				"type": "create",
				"url": "https://orders/1"
			}]
		`)
		if want := http.StatusInternalServerError; r.response.StatusCode != want {
			t.Fatalf("status: got %d, want %d", r.response.StatusCode, want)

		}
		if want, got := "500 Internal Server Error\n", r.readBody(); want != got {
			t.Errorf("body: got %q, want %q", got, want)
		}

		wantLog := "unknown error"
		if !strings.Contains(r.readLogs(), wantLog) {
			t.Errorf("logs: got %q, want %q", r.readLogs(), wantLog)
		}
	})
}

type testRunner struct {
	uuid      string
	error     bool
	panic     bool
	events    []*ReceivedEvent
	response  *http.Response
	logBuffer bytes.Buffer
}

func newTestRunner(uuid string) *testRunner {
	return &testRunner{uuid: uuid}
}

func (t *testRunner) do(url, username, body string) {
	logger := log.New(&t.logBuffer, "", 0)
	listener := NewListener(&ListenerConfig{
		Handler: func(events []*ReceivedEvent) error {
			if t.panic {
				panic(errors.New("panic error"))
			}
			if t.error {
				return errors.New("unknown error")
			}
			t.events = events
			return nil
		},
		Logger: logger,
		OnError: func(err error) {
			logger.Println(err.Error())
		},
		UUID: t.uuid,
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

func (t *testRunner) readLogs() string {
	result := t.logBuffer.String()
	t.logBuffer = bytes.Buffer{}
	return result
}
