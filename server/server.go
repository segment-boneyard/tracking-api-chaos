package server

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gohttp/app"
	"github.com/gohttp/response"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"github.com/segmentio/events"
	"github.com/segmentio/tracking-api-chaos/message"
	"github.com/segmentio/tracking-api-chaos/tracker"
)

// CORS options.
var options = cors.Options{
	AllowedOrigins:   []string{"*"},
	AllowedMethods:   []string{"POST", "OPTIONS"},
	AllowedHeaders:   []string{"*"},
	AllowCredentials: true,
	MaxAge:           604800,
}

// Channel.
const channel = "server"

// 500kb Limit that should apply to all routes. Same as the batch limit as that
// is our biggest limit.
const limit int64 = 500 << 10

// Response
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// Routes.
var Routes = map[string]string{
	"/v1/identify": "identify",
	"/v1/group":    "group",
	"/v1/alias":    "alias",
	"/v1/page":     "page",
	"/v1/screen":   "screen",
	"/v1/track":    "track",
	"/v1/batch":    "batch",
	"/v1/import":   "batch",
}

// Server structure.
type Server struct {
	tracker *tracker.Tracker
	*app.App
}

// New returns a new Server.
func New(t *tracker.Tracker) http.Handler {
	srv := &Server{tracker: t, App: app.New()}

	for route := range Routes {
		srv.Post(route, srv.handle)
		srv.Put(route, srv.handle)
	}

	return cors.New(options).Handler(srv)
}

// Handle requests coming from server side libraries.
func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	typ := Routes[r.URL.Path]
	encoding := strings.TrimSpace(r.Header.Get("Content-Encoding"))

	if encoding == "gzip" {
		z, err := gzip.NewReader(r.Body)
		if err != nil {
			events.Log("[server]: %{error}s", errors.Wrap(err, "gzip reader error"))
			response.BadRequest(w, &Response{
				Success: false,
				Message: "Malformed gzip content",
			})
			return
		}
		defer z.Close()

		r.Body = z
	}

	// Read the body now since if the request errors, we can't read it after
	// `message.FromRequest`. Limit the reader so we don't try to read too much.
	limitReader := http.MaxBytesReader(w, r.Body, limit)
	b, err := ioutil.ReadAll(limitReader)
	if err != nil {
		response.BadRequest(w)
		return
	}
	r.Body = ioutil.NopCloser(bytes.NewReader(b))

	// We support Content-Encoding as an extension to the HTTP standard in order
	// to receive compressed payloads, so we have to strip it out, otherwise it
	// would get passed upstream and be misleading since the content is decoded
	// by tracking-api already.
	r.Header.Del("Content-Encoding")

	msg, err := message.FromRequest(typ, w, r)
	if err != nil {
		// Most errors are connections being dropped and the JSON decoder returning
		// "unexpected EOF", this is not valuable information so we don't log it.
		if err != io.ErrUnexpectedEOF {
			events.Log("[server]: %{error}s", errors.Wrap(err, "reading request body"))
		}
		response.BadRequest(w)
		return
	}

	if err := s.tracker.Publish(ctx, msg); err != nil {
		response.InternalServerError(w)
		return
	}

	response.JSON(w, &Response{
		Success: true,
	})
}
