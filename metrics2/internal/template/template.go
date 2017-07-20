package template

import (
	"regexp"
	"strings"

	metrics "github.com/go-kit/kit/metrics2"
)

var templateRegexp = regexp.MustCompile(`{([^{}]+)}`)

// ExtractKeysFrom a templated name.
// For example, "foo_{x}_{y}_bar" will yield {"x", "y"}.
func ExtractKeysFrom(tmpl string) []string {
	keys := []string{}
	for _, match := range templateRegexp.FindAllStringSubmatch(tmpl, -1) {
		keys = append(keys, strings.Trim(match[0], "{}"))

	}
	return keys
}

// Render a templated name like "foo_{x}_{y}_bar" to "foo_abc_unknown_bar".
func Render(tmpl string, fields map[string]string) string {
	for _, match := range templateRegexp.FindAllStringSubmatch(tmpl, -1) {
		tok, key, val := match[0], match[1], metrics.UnknownValue
		if field, ok := fields[key]; ok {
			val = field
		}
		tmpl = strings.Replace(tmpl, tok, val, 1)
	}
	return tmpl
}
