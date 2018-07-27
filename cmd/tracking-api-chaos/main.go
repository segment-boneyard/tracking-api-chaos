package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/segmentio/conf"
	"github.com/segmentio/events"
	_ "github.com/segmentio/events/ecslogs"
	"github.com/segmentio/events/httpevents"
	_ "github.com/segmentio/events/log"
	_ "github.com/segmentio/events/text"
	"github.com/segmentio/tracking-api-chaos/api"
	"github.com/segmentio/tracking-api-chaos/chaos"
	yaml "gopkg.in/yaml.v2"
)

type config struct {
	Bind            string        `conf:"bind"             help:"Address on which tracking-api listens for incoming connections." validate:"nonzero"`
	Debug           bool          `conf:"debug"            help:"Turn on debug mode."`
	Out             string        `conf:"out" help:"file to write requests to (default: /dev/null)"`
	ErrorsOut       string        `conf:"errors-out" help:"file to write errors to (default: /dev/null)"`
	ShutdownTimeout time.Duration `conf:"shutdown-timeout" help:"Time limit for shutting down tracking-api." validate:"min=5"`
}

var version = "dev"

func main() {
	config := config{
		Bind:      ":3000",
		Out:       "/dev/null",
		ErrorsOut: "/dev/null",
	}
	conf.Load(&config)
	events.DefaultLogger.EnableDebug = config.Debug
	events.DefaultLogger.EnableSource = config.Debug

	var chaosRoot chaos.WeightedChaos
	err := yaml.Unmarshal([]byte(chaos.YamlConfigExample), &chaosRoot)
	if err != nil {
		events.Log("unmarshalling chaoses failed: %{error}s", err)
		os.Exit(1)
	}

	events.Log("starting %s, version: %s", os.Args[0], version)
	events.Debug("chaosRoot: %#v", chaosRoot)

	var handler http.Handler
	handler = api.New(config.Out, config.ErrorsOut, chaosRoot)

	if config.Debug {
		handler = httpevents.NewHandler(handler)
	}

	lstn, err := net.Listen("tcp", config.Bind)
	if err != nil {
		events.Log("binding %{address}s failed: %{error}s", config.Bind, err)
		os.Exit(1)
	}
	defer lstn.Close()

	sigsend := make(chan os.Signal)
	sigrecv := events.Signal(sigsend)
	signal.Notify(sigsend, syscall.SIGINT, syscall.SIGTERM)

	server := http.Server{
		Handler: handler,
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		<-sigrecv

		// Remove the handlers so if the signal is sent a second time we force
		// the termination of the program.
		signal.Stop(sigsend)

		ctx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
		defer cancel()
		server.Shutdown(ctx)

		if ctx.Err() != nil {
			events.Log("shutting down the http server is taking too long, giving up after %{shutdown_timeout}s: %{err}s",
				config.ShutdownTimeout,
				ctx.Err(),
			)
			server.Close()
		}
	}()

	events.Log("serving requests on %{bind_address}s", config.Bind)

	exitCode := 0
	switch err := server.Serve(lstn); err {
	case http.ErrServerClosed:
		events.Log("waiting for the http server to shut down")
		// On a clean shutdown we wait for the server to be terminate.
		wg.Wait()
	default:
		exitCode = 1
		events.Log("an error occured serving requests: %{error}v", err)
	}

	os.Exit(exitCode)
}
