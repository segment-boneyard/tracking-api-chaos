package tracker

import (
	"encoding/json"
	"io"
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
	out     io.Writer
	outJson *json.Encoder

	// We lock our output files to ensure no overlapping writes
	// TODO: is this necessary? leaning yes
	outLock sync.Mutex
}

// New returns a new tracker.
func New(out io.Writer) *Tracker {
	return &Tracker{
		out:     out,
		outJson: json.NewEncoder(out),
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
	events.Log("[tracker]: %{error}s", errors.Wrap(err, "marshaling JSON"))
	return
}
