package pixel

import (
	"net/http"

	"github.com/gohttp/app"
	"github.com/pkg/errors"
	"github.com/segmentio/events"
	"github.com/segmentio/tracking-api-chaos/message"
	"github.com/segmentio/tracking-api-chaos/tracker"
)

// GIF is an empty transparent gif.
var GIF = []byte{71, 73, 70, 56, 57, 97, 1, 0, 1, 0, 128, 0, 0, 0, 0, 0, 255, 255, 255, 33, 249, 4, 1, 0, 0, 0, 0, 44, 0, 0, 0, 0, 1, 0, 1, 0, 0, 2, 1, 68, 0, 59}

// Routes.
var Routes = map[string]string{
	"/v1/pixel/identify": "identify",
	"/v1/pixel/group":    "group",
	"/v1/pixel/alias":    "alias",
	"/v1/pixel/page":     "page",
	"/v1/pixel/screen":   "screen",
	"/v1/pixel/track":    "track",
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
	}

	return srv
}

// Handle handles requests.
func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	typ := Routes[r.URL.Path]

	var msg *message.Message
	var err error

	if data := r.URL.Query().Get("data"); len(data) > 0 {
		msg, err = message.FromBase64(typ, r)
	} else {
		msg, err = message.FromQuery(typ, r)
	}

	if err != nil {
		events.Log("[pixel]: %{error}s", errors.Wrap(err, "reading pixel data"))
	} else {
		s.tracker.Publish(r.Context(), msg)
	}

	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-cache, max-age=0")
	w.WriteHeader(200)
	w.Write(GIF)
}
