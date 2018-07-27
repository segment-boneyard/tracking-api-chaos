package message

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

const (
	propertyReceivedAt = "receivedAt"
)

// Payload is the payload interface.
type Payload interface {
	// SetReceivedAt sets the received at property on the payload
	SetReceivedAt(t time.Time) error

	// ClearProperty removes the specified top level property
	ClearProperty(name string) (ok bool)
}

// RawBody is used for decoding json.
type RawBody map[string]*json.RawMessage

// SetReceivedAt implementation.
func (rb RawBody) SetReceivedAt(t time.Time) error {
	raw, err := json.Marshal(t)
	if err != nil {
		return err
	}

	time := json.RawMessage(raw)
	rb[propertyReceivedAt] = &time
	return nil
}

func (rb RawBody) ClearProperty(name string) bool {
	_, ok := rb[name]
	delete(rb, name)
	return ok
}

// Body is used for manually constructing messages.
type Body map[string]interface{}

// SetReceivedAt implementation.
func (b Body) SetReceivedAt(t time.Time) error {
	b[propertyReceivedAt] = t
	return nil
}

func (b Body) ClearProperty(name string) bool {
	_, ok := b[name]
	delete(b, name)
	return ok
}

// Message structure.
type Message struct {
	Body    Payload     `json:"body"`    // Request body
	Method  string      `json:"method"`  // Request method
	Path    string      `json:"path"`    // Request path
	Headers http.Header `json:"headers"` // Request headers
}

// New creates a new Message.
func New(r *http.Request) *Message {
	return &Message{
		Method:  r.Method,
		Path:    r.URL.Path,
		Headers: makeMessageHeader(r.Header),
	}
}

// Batch returns true if the message is batch.
func (m *Message) Batch() bool {
	return "/v1/import" == m.Path ||
		"/v1/batch" == m.Path ||
		"/v1/b" == m.Path
}

func makeMessageHeader(requestHeader http.Header) http.Header {
	messageHeader := make(http.Header, len(requestHeader))

	for name, values := range requestHeader {
		if !isHopByHopHeader(name) {
			messageHeader[name] = values
		}
	}

	return messageHeader
}

func isHopByHopHeader(name string) bool {
	// http://freesoft.org/CIE/RFC/2068/143.htm
	for _, header := range [...]string{
		"Connection",
		"Keep-Alive",
		"Public",
		"Proxy-Authenticate",
		"Transfer-Encoding",
		"Upgrade",
	} {
		if strings.EqualFold(name, header) {
			return true
		}
	}
	return false
}
