package appdash

import (
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/appdash"
)

func TestNewEndpointEvent(t *testing.T) {
	e := DefaultEndpointEvent{}
	if want, have := "Endpoint", e.Schema(); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
	as, err := appdash.MarshalEvent(e)
	if err != nil {
		t.Fatal(err)
	}
	if want, have := map[string]string{
		"_schema:Endpoint": "",
		"Endpoint.Name":    "",
		"Endpoint.Recv":    "0001-01-01T00:00:00Z",
		"Endpoint.Send":    "0001-01-01T00:00:00Z",
		"Endpoint.Err":     "",
	}, as.StringMap(); !reflect.DeepEqual(want, have) {
		t.Errorf("want %#v, have %#v", want, have)
	}
}
