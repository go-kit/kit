package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/go-kit/kit/metrics2"
	"github.com/go-kit/kit/metrics2/expvar"
)

func main() {
	// Build the flags.
	fs := flag.NewFlagSet("main", flag.ExitOnError)
	var (
		listen = fs.String("listen", ":8080", "HTTP listen address")
	)
	fs.Usage = usageFor(fs)
	fs.Parse(os.Args[1:])

	// Build the provider and metrics.
	var (
		provider       metrics.Provider
		requestCount   metrics.Counter
		requestLatency metrics.Histogram
	)
	{
		provider = expvar.NewProvider()
		requestCount = provider.NewCounter(metrics.Identifier{
			NameTemplate: "http_request_{method}_{code}_count",
		})
		requestLatency = provider.NewHistogram(metrics.Identifier{
			NameTemplate: "http_request_{method}_{code}_seconds",
		})
	}

	// An instrumenting middleware, to update the metrics.
	instrument := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			begin := time.Now()
			iw := &interceptingWriter{ResponseWriter: w, code: http.StatusOK}
			next.ServeHTTP(iw, r)
			requestCount.With("method", r.Method, "code", strconv.Itoa(iw.code)).Add(1)
			requestLatency.With("method", r.Method, "code", strconv.Itoa(iw.code)).Observe(time.Since(begin).Seconds())
		})
	}

	// The business logic HTTP handler.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := http.StatusOK
		if i, err := strconv.Atoi(r.FormValue("code")); err == nil {
			code = i
		}
		log.Printf("%s: %s %s -> %d", r.RemoteAddr, r.Method, r.URL.String(), code)
		w.WriteHeader(code)
	})

	// Launch the HTTP server.
	http.Handle("/", instrument(handler))
	log.Printf("listening on %s", *listen)
	log.Fatal(http.ListenAndServe(*listen, nil))
}

type interceptingWriter struct {
	http.ResponseWriter
	code int
}

func (iw *interceptingWriter) WriteHeader(code int) {
	iw.code = code
	iw.ResponseWriter.WriteHeader(code)
}

func usageFor(fs *flag.FlagSet) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "USAGE\n")
		fmt.Fprintf(os.Stderr, "  demo [flags]\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "FLAGS\n")
		w := tabwriter.NewWriter(os.Stderr, 0, 4, 2, ' ', 0)
		fs.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(w, "  -%s %s\t%s\n", f.Name, f.DefValue, f.Usage)
		})
		w.Flush()
		fmt.Fprintf(os.Stderr, "\n")

	}
}
