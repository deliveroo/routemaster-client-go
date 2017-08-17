package routemaster

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestValidatesParams(t *testing.T) {
	tests := []struct {
		name, url, uuid string
		valid           bool
	}{
		{"empty url", "", "demo", false},
		{"invalid url", "not a url", "demo", false},
		{"empty uuid", "https://routemaster.dev", "", false},
		{"valid params", "https://routemaster.dev", "demo", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(&Config{URL: tt.url, UUID: tt.uuid})
			if tt.valid && (err != nil) {
				t.Error("unexpected validation error")
			} else if !tt.valid && (err == nil) {
				t.Error("expected validation error")
			}
		})
	}
}

func TestRequests(t *testing.T) {
	assertEqJSON := func(field string, want, got interface{}) {
		wantJSON := prettyJSON(want)
		gotJSON := prettyJSON(got)
		if gotJSON != wantJSON {
			t.Errorf("incorrect %s:\nwant:\n%+v\n\ngot:\n%+v\n", field, wantJSON, gotJSON)
		}
	}
	assertDeepEq := func(field string, want, got interface{}) {
		if !reflect.DeepEqual(want, got) {
			t.Errorf("incorrect %s:\nwant:\n%s\n\ngot:\n%s\n", field, want, got)
		}
	}
	tests := []struct {
		name     string
		run      func(*Client) error
		method   string
		username string
		path     string
		body     interface{}
		resp     interface{}
	}{
		{
			name: "Get Token",
			run: func(c *Client) error {
				tokens, err := c.GetTokens()
				assertDeepEq("tokens", tokens, []*TokenResponse{
					{Name: "client", Token: "client-123"},
				})
				return err
			},
			method:   http.MethodGet,
			path:     "/api_tokens",
			username: "root",
			body:     nil,
			resp: []M{
				{
					"name":  "client",
					"token": "client-123",
				},
			},
		},
		{
			name: "Create Token",
			run: func(c *Client) error {
				_, err := c.CreateToken("client")
				return err
			},
			method:   http.MethodPost,
			path:     "/api_tokens",
			username: "root",
			body: M{
				"name": "client",
			},
			resp: M{
				"name":  "client",
				"token": "client-123",
			},
		},
		{
			name: "Delete Token",
			run: func(c *Client) error {
				return c.DeleteToken("client")
			},
			method:   http.MethodDelete,
			path:     "/api_tokens/client",
			username: "root",
			body:     nil,
		},
		{
			name: "Push",
			run: func(c *Client) error {
				return c.Push(&Event{
					Topic: "orders",
					Type:  "create",
					URL:   "https://orders/1",
				})
			},
			method:   http.MethodPost,
			path:     "/topics/orders",
			username: "demo",
			body: M{
				"type": "create",
				"url":  "https://orders/1",
			},
		},
		{
			name: "Subscribe",
			run: func(c *Client) error {
				return c.Subscribe(&Subscription{
					Topics:   []string{"orders"},
					Callback: "https://localhost",
					UUID:     "demo",
					Timeout:  500,
					Max:      0,
				})
			},
			method:   http.MethodPost,
			path:     "/subscription",
			username: "demo",
			body: M{
				"topics":   []string{"orders"},
				"uuid":     "demo",
				"callback": "https://localhost",
				"timeout":  500,
			},
		},
		{
			name: "Delete Topic",
			run: func(c *Client) error {
				return c.DeleteTopic("orders")
			},
			method:   http.MethodDelete,
			path:     "/topic/orders",
			username: "demo",
			body:     nil,
		},
		{
			name: "Unsubscribe",
			run: func(c *Client) error {
				return c.Unsubscribe("orders")
			},
			method:   http.MethodDelete,
			path:     "/subscriber/topics/orders",
			username: "demo",
			body:     nil,
		},
		{
			name: "Unsubscribe All",
			run: func(c *Client) error {
				return c.UnsubscribeAll()
			},
			method:   http.MethodDelete,
			path:     "/subscriber",
			username: "demo",
			body:     nil,
		},
		{
			name: "Get Topics",
			run: func(c *Client) error {
				topics, err := c.GetTopics()
				assertDeepEq("topics", topics, []*Topic{
					{Name: "orders", Publisher: "demo", Events: 1},
				})
				return err
			},
			method:   http.MethodGet,
			path:     "/topics",
			username: "demo",
			resp: []M{
				{"name": "orders", "publisher": "demo", "events": 1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				username, _, _ := r.BasicAuth()

				if r.URL.Path != tt.path {
					t.Errorf("incorrect path: want: %s, got: %s\n", tt.path, r.URL.Path)
				}
				if r.Method != tt.method {
					t.Errorf("incorrect http method: want: %s, got %s\n", tt.method, r.Method)
				}
				if username != tt.username {
					t.Errorf("incorrect username: want: %s, got: %s\n", tt.username, username)
				}
				assertEqJSON("body", tt.body, readJSON(r.Body))
				if tt.resp != nil {
					resp, _ := json.Marshal(tt.resp)
					w.Write(resp)
				}
			}))
			defer ts.Close()

			client, err := NewClient(&Config{URL: ts.URL, UUID: tt.username})
			must(err)

			must(tt.run(client))
		})
	}

}
