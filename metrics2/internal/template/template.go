package template

import (
	"regexp"
	"strings"
)

var (
	templateRegexp = regexp.MustCompile(`{([^{}]+)}`)
	defaultValue   = "unknown"
)

// Render a templated name like "foo_{x}_{y}_bar" to "foo_abc_unknown_bar".
func Render(s string, fields map[string]string) string {
	for _, match := range templateRegexp.FindAllStringSubmatch(s, -1) {
		tok, key, val := match[0], match[1], defaultValue
		if field, ok := fields[key]; ok {
			val = field
		}
		s = strings.Replace(s, tok, val, 1)
	}
	return s
}
