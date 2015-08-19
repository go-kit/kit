package appdash

import (
	"reflect"
	"sourcegraph.com/sourcegraph/appdash"
	"testing"
)

func TestNewEndpointEvent(t *testing.T) {
	e := NewDefaultEndpointEvent()

	if e.Schema() != "Endpoint" {
		t.Errorf("unexpected schema: %v", e.Schema())
	}

	anns, err := appdash.MarshalEvent(e)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]string{
		"_schema:Endpoint": "",
		"Endpoint.Name":    "",
		"Endpoint.Recv":    "0001-01-01T00:00:00Z",
		"Endpoint.Send":    "0001-01-01T00:00:00Z",
		"Endpoint.Err":     "",
	}

	if !reflect.DeepEqual(anns.StringMap(), expected) {
		t.Errorf("got %#v, want %$v", anns.StringMap(), expected)
	}
}
