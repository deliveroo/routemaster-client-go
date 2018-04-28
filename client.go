package routemaster

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// M is shorthand for map[string]interface{}.
type M map[string]interface{}

// Config specifies the parameters needed to instantiate a Client.
type Config struct {
	// IgnoreSSL disables SSL verification for URL.
	IgnoreSSL bool

	// URL is the URL of the Routemaster bus.
	URL string

	// UUID is the unique client identifier.
	UUID string
}

func (c *Config) validate() error {
	if c.URL == "" {
		return errors.New("URL must not be empty")
	}
	if !isValidAbsoluteURL(c.URL) {
		return errors.New("URL must be a valid absolute url")
	}
	if c.UUID == "" {
		return errors.New("UUID most not be empty")
	}
	return nil
}

// Client is a Routemaster API client.
type Client struct {
	config *Config
	client *http.Client
}

// NewClient instantiates a new Routemaster API client.
func NewClient(config *Config) (*Client, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}

	client := http.DefaultClient
	if config.IgnoreSSL {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	}
	return &Client{
		config: config,
		client: client,
	}, nil
}

func (c *Client) do(method, path string, body interface{}, result interface{}) error {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(method,
		c.config.URL+path,
		bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.config.UUID, "")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	if !isHTTPSuccess(resp.StatusCode) {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		return &clientError{
			status:      resp.Status,
			respHeaders: resp.Header,
			respBody:    body,
			reqBody:     bodyBytes,
		}
	}
	if result != nil {
		defer resp.Body.Close()
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(buf, &result); err != nil {
			return err
		}
	}
	return nil
}

type clientError struct {
	status      string
	statusCode  int
	respHeaders http.Header
	respBody    []byte
	reqBody     []byte
}

func (e *clientError) ResponseHeaders() http.Header {
	return e.respHeaders
}

func (e *clientError) StatusCode() int {
	return e.statusCode
}

func (e *clientError) RequestBody() []byte {
	return e.reqBody
}

func (e *clientError) ResponseBody() []byte {
	return e.respBody
}

func (e *clientError) Error() string {
	return fmt.Sprintf("routemaster/client: status=%s response=%s request=%s respHeaders=%v",
		e.status,
		string(e.respBody),
		string(e.reqBody),
		e.ResponseHeaders(),
	)
}

func isHTTPSuccess(statusCode int) bool {
	switch statusCode {
	case http.StatusOK, http.StatusCreated, http.StatusNoContent:
		return true
	}
	return false
}

// GetTokens retrieves all registered API tokens.
func (c *Client) GetTokens() ([]*Token, error) {
	var response []*Token
	err := c.do(http.MethodGet, "/api_tokens", nil, &response)
	return response, err
}

// CreateToken creates an API token.
func (c *Client) CreateToken(name string) (string, error) {
	var response Token
	err := c.do(http.MethodPost, "/api_tokens", M{"name": name}, &response)
	return response.Token, err
}

// DeleteToken deletes an API token.
func (c *Client) DeleteToken(token string) error {
	path := fmt.Sprintf("/api_tokens/%s", token)
	return c.do(http.MethodDelete, path, nil, nil)
}

// Subscribe subscribes a listener to a Routemaster topic.
func (c *Client) Subscribe(s *Subscription) error {
	if err := s.validate(); err != nil {
		return err
	}
	return c.do(http.MethodPost, "/subscription", s, nil)
}

// DeleteTopic deletes the specified topic.
func (c *Client) DeleteTopic(topic string) error {
	path := fmt.Sprintf("/topic/%s", topic)
	return c.do(http.MethodDelete, path, nil, nil)
}

// Push pushes an event to the Routemaster bus.
func (c *Client) Push(topic string, e *Event) error {
	if err := e.validate(); err != nil {
		return err
	}
	path := fmt.Sprintf("/topics/%s", topic)
	return c.do(http.MethodPost, path, e, nil)
}

// Unsubscribe unsubscribes a listener from a Routemaster topic.
func (c *Client) Unsubscribe(topic string) error {
	path := fmt.Sprintf("/subscriber/topics/%s", topic)
	return c.do(http.MethodDelete, path, nil, nil)
}

// UnsubscribeAll unsubscribes a listener from all topics.
func (c *Client) UnsubscribeAll() error {
	return c.do(http.MethodDelete, "/subscriber", nil, nil)
}

// GetTopics retrieves all topics.
func (c *Client) GetTopics() ([]*Topic, error) {
	var result []*Topic
	err := c.do(http.MethodGet, "/topics", nil, &result)
	return result, err
}
