package chaos

import (
	"net/http"
	"testing"

	"github.com/mitchellh/mapstructure"
)

type NamedChaos string

func (c NamedChaos) Do(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
	return w, r
}

func TestChoose(t *testing.T) {
	nc0_5 := NamedChaos("[0,5)")
	nc5_15 := NamedChaos("[5,15)")
	nc15_30 := NamedChaos("[15,30)")
	wc := WeightedChaos{
		{
			Weight: 5,
			Chaos:  &nc0_5,
		},
		{
			Weight: 10,
			Chaos:  &nc5_15,
		},
		{
			Weight: 15,
			Chaos:  &nc15_30,
		},
	}

	if wc.Choose(0) != &nc0_5 {
		t.Fail()
	}
	if wc.Choose(2.5) != &nc0_5 {
		t.Fail()
	}
	if wc.Choose(5) != &nc5_15 {
		t.Fail()
	}
	if wc.Choose(30) != nil {
		t.Fail()
	}
	if wc.Choose(30.1) != nil {
		t.Fail()
	}
}

func TestDecode(t *testing.T) {
	var chaos StatusCodeChaos
	chaos = StatusCodeChaos{}
	v := map[string]interface{}{
		"code": 500,
	}
	mapstructure.Decode(v, &chaos)
	if chaos.Code != 500 {
		t.Fail()
	}
}
