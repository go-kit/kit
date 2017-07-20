package keyval

import metrics "github.com/go-kit/kit/metrics2"

func MakeWith(keys []string) map[string]string {
	m := make(map[string]string, len(keys))
	for _, key := range keys {
		m[key] = metrics.UnknownValue
	}
	return m
}

// Merge the keyvals into the original map, and return a new map. Keyvals that
// aren't present in the original map are dropped.
func Merge(original map[string]string, keyvals ...string) map[string]string {
	if len(keyvals)%2 != 0 {
		keyvals = append(keyvals, metrics.UnknownValue)
	}
	result := map[string]string{}
	for k, v := range original {
		result[k] = v
	}
	for i := 0; i < len(keyvals); i += 2 {
		if _, ok := result[keyvals[i]]; ok {
			result[keyvals[i]] = keyvals[i+1]
		}
	}
	return result
}
