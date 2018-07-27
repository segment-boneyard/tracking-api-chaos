package message

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Single message is limited to 32KB.
const Single int64 = 32 << 10

// Batch message is limited to 500KB.
const Batch int64 = 500 << 10

// FromRequest reads a message `typ` from the given `request`.
func FromRequest(typ string, w http.ResponseWriter, r *http.Request) (*Message, error) {
	limit := Limit(typ)

	if "GET" == r.Method {
		return FromBase64(typ, r)
	}

	var body RawBody
	msg := New(r)
	req := http.MaxBytesReader(w, r.Body, limit)
	dec := json.NewDecoder(req)
	err := dec.Decode(&body)

	switch err {
	case nil:
	case io.ErrUnexpectedEOF:
		return nil, err
	default:
		return nil, fmt.Errorf("[message] error decoding json from request: %v", err)
	}

	msg.Body = body
	return msg, nil
}

// FromBase64 decodes data from `?data` query string.
func FromBase64(typ string, r *http.Request) (*Message, error) {
	limit := Limit(typ)
	data := r.URL.Query().Get("data")

	buf, err := decodeBase64(data)
	if err != nil {
		return nil, fmt.Errorf("[message] error decoding base64: %s (%q)", err, data)
	}

	if limit < int64(len(buf)) {
		return nil, fmt.Errorf("[message] request too large (limit=%v size=%v)", limit, len(buf))
	}

	var body RawBody
	msg := New(r)
	if err := json.Unmarshal(buf, &body); err != nil {
		return nil, fmt.Errorf("[message] error decoding json: %s (%q)", err, buf)
	}

	msg.Body = body
	return msg, nil
}

// FromQuery parses the query string and returns a message.
func FromQuery(typ string, r *http.Request) (*Message, error) {
	body := make(Body)
	values := r.URL.Query()

	for key, value := range values {
		parts := split(key)
		var val interface{}

		// ignore empty
		if len(parts) == 0 {
			continue
		}

		// consolidate value, respecting arrays
		if len(value) == 1 {
			val = value[0]
		} else {
			val = value
		}

		// single key
		if len(parts) == 1 {
			body[key] = val
			continue
		}

		// add sub keys
		keys := parts[0 : len(parts)-1]
		last := parts[len(parts)-1]
		ctx := body

		for _, key := range keys {
			if c, ok := ctx[key].(Body); ok {
				ctx = c
			} else {
				c := make(Body)
				ctx[key] = c
				ctx = c
			}
		}

		ctx[last] = val
	}

	msg := New(r)
	msg.Body = body
	return msg, nil
}

var base64Decoders = [...](func(string) ([]byte, error)){
	base64.RawURLEncoding.DecodeString,
	base64.URLEncoding.DecodeString,
	base64.RawStdEncoding.DecodeString,
	base64.StdEncoding.DecodeString,
}

// Decode base64 with support for Std and URL encodings.
func decodeBase64(data string) (b []byte, err error) {
	for _, d := range base64Decoders {
		if b, err = d(data); err == nil {
			return
		}
	}
	return
}

// Limit returns size limit for message `typ`.
func Limit(typ string) int64 {
	switch typ {
	case "batch":
		return Batch
	default:
		return Single
	}
}

// Split splits and trims elements at `.`.
// If any of the parts is empty an empty slice is returned.
func split(s string) []string {
	parts := strings.Split(s, ".")

	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil
		}

		parts[i] = part
	}

	return parts
}
