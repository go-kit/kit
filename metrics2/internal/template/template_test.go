package template

import (
	"reflect"
	"testing"
)

func TestExtractKeysFrom(t *testing.T) {
	for input, want := range map[string][]string{
		"":                []string{},
		"foo_bar":         []string{},
		"foo_{x}_bar":     []string{"x"},
		"foo_{x}_bar_{y}": []string{"x", "y"},
	} {
		t.Run(input, func(t *testing.T) {
			if have := ExtractKeysFrom(input); !reflect.DeepEqual(want, have) {
				t.Errorf("want %v, have %v", want, have)
			}
		})
	}
}

func TestRender(t *testing.T) {
	for _, testcase := range []struct {
		tmpl   string
		fields map[string]string
		want   string
	}{
		{"", map[string]string{}, ""},
		{"foo_bar", map[string]string{}, "foo_bar"},
		{"foo_{x}_bar", map[string]string{}, "foo_unknown_bar"},
		{"foo_{x}_bar", map[string]string{"y": "y"}, "foo_unknown_bar"},
		{"foo_{x}_bar", map[string]string{"x": "xxx"}, "foo_xxx_bar"},
		{"foo_{x}_bar_{y}", map[string]string{"x": "1"}, "foo_1_bar_unknown"},
		{"foo_{x}_bar_{y}", map[string]string{"x": "!", "y": "@"}, "foo_!_bar_@"},
	} {
		t.Run(testcase.tmpl, func(t *testing.T) {
			if want, have := testcase.want, Render(testcase.tmpl, testcase.fields); want != have {
				t.Errorf("want %q, have %q", want, have)
			}
		})
	}
}
