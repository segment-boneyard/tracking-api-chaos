package chaos

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/mitchellh/mapstructure"
	"github.com/segmentio/events"
)

// NB: no tabs here
const DefaultConfigYAML = `
- weight: 5
  latency:
    latency: 10000
- weight: 5
  latency:
    latency: 31000
- weight: 5
  statusCode:
    code: 500
    body: "Something went wrong"
- weight: 5
  statusCode:
    code: 429
    body: "slow down"
`

type FakeResponseWriter struct {
}

func (f *FakeResponseWriter) Header() http.Header {
	return make(http.Header)
}

func (f *FakeResponseWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

func (f *FakeResponseWriter) WriteHeader(statusCode int) {
}

type Kind string

const (
	StatusCode Kind = "statusCode"
	Latency    Kind = "latency"
)

type Chaos interface {
	Do(http.ResponseWriter, *http.Request) (http.ResponseWriter, *http.Request)
	// TODO: interrupt?
}

// Write a specific HTTP status code and body
type StatusCodeChaos struct {
	Code int    `mapstructure:"code"`
	Body []byte `mapstructure:"body"`
}

// If Body is not nil, w will be replaced with a FakeResponseWriter
func (c StatusCodeChaos) Do(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
	w.WriteHeader(c.Code)
	if c.Body == nil {
		return w, r
	}
	w.Write(c.Body)

	return &FakeResponseWriter{}, r
}

// Delay request by some amount. Amount is `Latency` ms plus or minus a random amount
// of jitter, up to `Jitter` ms
type LatencyChaos struct {
	Latency int64 `mapstructure:"latency"`
	Jitter  int64 `mapstructure:"jitter"`
}

func (c LatencyChaos) delay() {
	delay := c.Latency
	jitter := c.Jitter
	if jitter > 0 {
		delay = rand.Int63n(jitter*2) - jitter
	}
	// TODO: this is blocking; do we need a way to interrupt?
	time.Sleep(time.Duration(delay) * time.Millisecond)
}

func (c LatencyChaos) Do(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
	c.delay()
	return w, r
}

const DefaultWeight float64 = 100

type WeightedChaosItem struct {
	Weight float64
	Chaos  Chaos
}

type WeightedChaos []WeightedChaosItem

func (c *WeightedChaos) UnmarshalYAML(unmarshal func(interface{}) error) (err error) {
	var itemsmap []map[string]interface{}
	err = unmarshal(&itemsmap)
	if err != nil {
		events.Log("failed basic unmarshal: %{error}s", err)
		return err
	}
	var weightsum float64 = 0
	for _, item := range itemsmap {
		var weight float64
		weight_untyped, ok := item["weight"]
		if !ok {
			weight = DefaultWeight
		} else {
			weight, ok = weight_untyped.(float64)
			if !ok {
				weight_int, ok := weight_untyped.(int)
				if !ok {
					err = multierror.Append(err, fmt.Errorf("unable to interpret type %#v as float64 in item %s", weight_untyped, item))
					continue
				}
				weight = float64(weight_int)
			}
		}
		delete(item, "weight")
		if len(item) != 1 {
			err = multierror.Append(err, fmt.Errorf("item must have exactly 1 chaos; has %d: %s", len(item), item))
			continue
		}
		var k string
		var v interface{}
		// get the first and only item
		// TODO? better way to do this
		for k, v = range item {
		}
		var chaos Chaos
		kind := Kind(k)
		switch kind {
		case StatusCode:
			chaosTyped := StatusCodeChaos{}
			mapstructure.Decode(v, &chaosTyped)
			chaos = chaosTyped
		case Latency:
			chaosTyped := LatencyChaos{}
			mapstructure.Decode(v, &chaosTyped)
			chaos = chaosTyped
		default:
			err = multierror.Append(err, fmt.Errorf("unrecognized chaos type `%s`", k))
			continue
		}
		weightsum += weight
		*c = append(*c, WeightedChaosItem{
			Weight: weight,
			Chaos:  chaos,
		},
		)
	}
	if weightsum > 100 {
		err = multierror.Append(err, fmt.Errorf("sum of weights must be < 100; is %f", weightsum))
	}
	if err != nil {
		*c = nil
	}
	return err
}

func (c WeightedChaos) Choose(i float64) Chaos {
	for _, item := range c {
		if i < item.Weight {
			return item.Chaos
		}
		i -= item.Weight
	}
	return nil
}

func (c WeightedChaos) Do(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
	rander := rand.New(rand.NewSource(time.Now().UnixNano()))
	i := rander.Float64() * 100
	chaos := c.Choose(i)
	if chaos != nil {
		events.Debug("Causing chaos %#v", chaos)
		w, r = chaos.Do(w, r)
	}
	return w, r
}

type NopChaos struct{}

func (c NopChaos) Do(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
	return w, r
}
