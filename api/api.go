package api

import (
	"io"
	"net/http"
	"strings"

	"github.com/gohttp/app"
	"github.com/rs/cors"
	"github.com/segmentio/tracking-api-chaos/chaos"
	"github.com/segmentio/tracking-api-chaos/client"
	"github.com/segmentio/tracking-api-chaos/crossdomain"
	"github.com/segmentio/tracking-api-chaos/pixel"
	"github.com/segmentio/tracking-api-chaos/server"
	"github.com/segmentio/tracking-api-chaos/tracker"
)

type Server struct {
	pixel  http.Handler
	server http.Handler
	client http.Handler
	chaos  chaos.Chaos
	*app.App
}

func New(out, errorsOut io.Writer, chaosRoot chaos.Chaos) *Server {
	api := &Server{
		App:   app.New(),
		chaos: chaosRoot,
	}
	tracker := tracker.New(out, errorsOut)
	api.pixel = pixel.New(tracker)
	api.client = cors.Default().Handler(client.New(tracker))
	api.server = server.New(tracker)
	api.Use(api.route)
	api.Get("/internal/health", api.health)
	api.Get("/crossdomain.xml", crossdomain.Route)
	return api
}

// Route routes using `pkg.Routes` to each server.
func (s *Server) route(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.ToLower(r.URL.Path)
		path := r.URL.Path

		if path[len(path)-1] == '/' {
			r.URL.Path = path[0 : len(path)-1]
			path = r.URL.Path
		}

		var downstream http.Handler

		if _, ok := pixel.Routes[path]; ok {
			downstream = s.pixel
		} else if _, ok := server.Routes[path]; ok {
			downstream = s.server
		} else if _, ok := client.Routes[path]; ok {
			downstream = s.client
			s.client.ServeHTTP(w, r)
		} else {
			downstream = h
		}

		w, r = s.chaos.Do(w, r)

		downstream.ServeHTTP(w, r)
	})
}

// Health responds with `200` and an empty body.
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Length", "0")
	w.Header().Set("Content-Type", "text/plain")

	w.WriteHeader(http.StatusOK)
}

var ok = []byte(`OK`)
