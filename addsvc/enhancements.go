package main

import (
	"encoding/json"
	"io"
	"time"
)

func logging(w io.Writer, add Add) Add {
	return func(a, b int64) (v int64) {
		defer func(begin time.Time) {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"a":      a,
				"b":      b,
				"result": v,
				"took":   time.Since(begin),
			})
		}(time.Now())
		v = add(a, b)
		return
	}
}
