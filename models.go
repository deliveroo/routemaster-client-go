package routemaster

import (
	"errors"
	"net/url"
)

// Subscription is a subscription to a Routemaster topic.
type Subscription struct {
	// Topics are the topics covered by the subscription.
	Topics []string `json:"topics"`

	// Callback is the URL of the event listener.
	Callback string `json:"callback"`

	// UUID is the password to be passed to the callback.
	UUID string `json:"uuid"`

	// Timeout is the minimal interval between callback requests. Optional.
	Timeout int `json:"timeout,omitempty"`

	// Max is the maximum number of events sent in each batch.
	Max int `json:"max,omitempty"`
}

func (s *Subscription) validate() error {
	if len(s.Topics) == 0 {
		return errors.New("routemaster: at least one topic must be specified")
	}
	u, err := url.Parse(s.Callback)
	if err != nil {
		return errors.New("routemaster: callback must be a valid URL")
	}
	if !u.IsAbs() {
		return errors.New("routemaster: callback must be https")
	}
	if s.UUID == "" {
		return errors.New("routemaster: UUID must be non-empty")
	}
	return nil
}

// Event is a Routemaster event.
type Event struct {
	// Topic is the topic to which this event belongs.
	Topic string `json:"topic,omitempty"`

	// Type is one of create, update, delete, or noop.
	Type string `json:"type"`

	// URL is the authoritative URL of the entity corresponding to the event.
	URL string `json:"url"`

	// Timestamp is when the event occurerd. Optional.
	Timestamp int64 `json:"timestamp,omitempty"`

	// Data is the payload associated with the event. Optional, and its use is
	// discouraged.
	Data interface{} `json:"data,omitempty"`
}

func (e *Event) validate() error {
	if e.Topic == "" {
		return errors.New("routemaster: topic must be non-empty")
	}
	switch e.Type {
	case "create", "update", "delete", "noop":
	default:
		return errors.New("routemaster: type must be one of create, update, delete, noop")
	}
	u, err := url.Parse(e.URL)
	if err != nil {
		return errors.New("routemaster: URL must be a valid URL")
	}
	if !u.IsAbs() {
		return errors.New("routemaster: URL must be https")
	}
	return nil
}

// Token represents an API token.
type Token struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}

// Topic is a Routemaster topic.
type Topic struct {
	// Name of the topic.
	Name string `json:"name"`

	// Publisher of the topic.
	Publisher string `json:"publisher"`

	// Number of events ever sent on a given topic.
	Events int `json:"events"`
}
