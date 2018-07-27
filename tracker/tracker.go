package tracker

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"context"

	"github.com/pkg/errors"
	"github.com/segmentio/events"
	"github.com/segmentio/tracking-api-chaos/message"
)

var Now = func() time.Time {
	return time.Now().UTC()
}

type Tracker struct {
	out           io.Writer
	errorsOut     io.Writer
	outJson       *json.Encoder
	errorsOutJson *json.Encoder

	// We lock our output files to ensure no overlapping writes
	// TODO: is this necessary? leaning yes
	outLock       sync.Mutex
	errorsOutLock sync.Mutex
}

type ErrorOut struct {
	Path  string
	Auth  string
	Error string
	Body  []byte
}

// New returns a new tracker.
func New(outFn, errorsOutFn string) *Tracker {

	out, err := os.Create(outFn)
	if err != nil {
		events.Log("opening out %{out}s failed: %{error}s", outFn, err)
		os.Exit(1)
	}

	errorsOut, err := os.Create(errorsOutFn)
	if err != nil {
		events.Log("opening errors-out %{errorsout}s failed: %{error}s", errorsOutFn, err)
		os.Exit(1)
	}
	return &Tracker{
		out:           out,
		outJson:       json.NewEncoder(out),
		errorsOut:     errorsOut,
		errorsOutJson: json.NewEncoder(errorsOut),
	}
}

// Writes errors (a JSON-serialized stream of `ErrorOut`s) to s.errorsOut
func (t *Tracker) PublishError(ctx context.Context, b []byte, r *http.Request, e error) {
	user, password, _ := r.BasicAuth()
	errorOut := ErrorOut{
		Path:  r.URL.Path,
		Auth:  fmt.Sprintf("%s:%s", user, password),
		Error: e.Error(),
		Body:  b,
	}

	t.errorsOutLock.Lock()
	defer t.errorsOutLock.Unlock()

	err := t.errorsOutJson.Encode(errorOut)
	if err != nil {
		events.Log("[tracker]: %{error}s", errors.Wrap(err, "encoding error"))
	} else {
		fmt.Fprintf(t.errorsOut, "\n\n")
	}
}

// Writes a msg to s.out
func (t *Tracker) Publish(ctx context.Context, msg *message.Message) (err error) {

	if err = msg.Body.SetReceivedAt(Now()); err != nil {
		events.Log("[tracker]: %{error}s", errors.Wrap(err, "setting received time"))
		return
	}

	t.outLock.Lock()
	defer t.outLock.Unlock()

	err = t.outJson.Encode(msg)
	if err != nil {
		events.Log("[tracker]: %{error}s", errors.Wrap(err, "marshaling JSON"))
	} else {
		fmt.Fprintf(t.errorsOut, "\n\n")
	}
	return
}
