package client

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gohttp/app"
	"github.com/gohttp/response"
	"github.com/pkg/errors"
	"github.com/segmentio/events"
	"github.com/segmentio/tracking-api-chaos/message"
	"github.com/segmentio/tracking-api-chaos/tracker"
)

// Channel.
const channel = "client"

// Success response.
type Success struct {
	Success bool `json:"success"`
}

// Routes.
var Routes = map[string]string{
	"/v1/i": "identify",
	"/v1/g": "group",
	"/v1/a": "alias",
	"/v1/p": "page",
	"/v1/s": "screen",
	"/v1/t": "track",
	"/v1/b": "batch",
}

// Server structure.
type Server struct {
	tracker *tracker.Tracker
	*app.App
}

// New returns a new Server.
func New(t *tracker.Tracker) *Server {
	srv := &Server{tracker: t, App: app.New()}

	for route, _ := range Routes {
		srv.Get(route, srv.handle)
		srv.Post(route, srv.handle)
		srv.Put(route, srv.handle)
	}

	return srv
}

// Handle handles requests.
func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	typ := Routes[r.URL.Path]
	msg, err := message.FromRequest(typ, w, r)

	if err != nil {
		// Most errors are connections being dropped and the JSON decoder returning
		// "unexpected EOF", this is not valuable information so we don't log it.
		if err != io.ErrUnexpectedEOF {
			events.Log("[client]: %{method}s %{path}s: %{error}s", r.Method, r.URL.Path, errors.WithStack(err))
		}
	} else {
		s.tracker.Publish(r.Context(), msg)
	}

	// JSONP
	if name := query.Get("callback"); name != "" {
		w.Header().Set("Content-Type", "text/javascript")
		w.WriteHeader(200)
		fmt.Fprintf(w, `typeof %s == "function" && %s({ success: true });`, name, name)
		return
	}

	if r.Method == "GET" {
		w.Header().Set("Cache-Control", "no-cache, max-age=0")
	}

	// JSON.
	response.JSON(w, &Success{true})
}
