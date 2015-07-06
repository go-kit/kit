package http

import (
	"net/http"

	"golang.org/x/net/context"
)

// BeforeFunc TODO
type BeforeFunc func(context.Context, *http.Request) context.Context

// AfterFunc TODO
type AfterFunc func(context.Context, http.ResponseWriter)

// SetContentType TODO
func SetContentType(contentType string) AfterFunc {
	return SetHeader("Content-Type", contentType)
}

// SetHeader TODO
func SetHeader(key, val string) AfterFunc {
	return func(_ context.Context, w http.ResponseWriter) {
		w.Header().Set(key, val)
	}
}
