package main

import (
"flag"
"fmt"
"net/http"
"os"
"os/signal"
"syscall"

"github.com/prometheus/client_golang/prometheus/promhttp"

"github.com/go-kit/kit/log"
"github.com/go-kit/kit/examples/library/inmemory"
"github.com/go-kit/kit/examples/library/catalog"
)

const (
	defaultPort              = "9090"
)

func main() {
	var (
		addr  = envString("PORT", defaultPort)

		httpAddr          = flag.String("http.addr", ":"+addr, "HTTP listen address")
	)

	flag.Parse()

	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	var (
		books    = inmemory.NewBookRepository()
	)


	// Facilitate testing by adding some cargos.
	//storeTestData(cargos)

	//fieldKeys := []string{"method"}

	var cs catalog.Service
	cs = catalog.NewService(books)
	//bs = booking.NewLoggingService(log.With(logger, "component", "booking"), bs)
	//bs = booking.NewInstrumentingService(
	//	kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
	//		Namespace: "api",
	//		Subsystem: "booking_service",
	//		Name:      "request_count",
	//		Help:      "Number of requests received.",
	//	}, fieldKeys),
	//	kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
	//		Namespace: "api",
	//		Subsystem: "booking_service",
	//		Name:      "request_latency_microseconds",
	//		Help:      "Total duration of requests in microseconds.",
	//	}, fieldKeys),
	//	bs,
	//)

	httpLogger := log.With(logger, "component", "http")

	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./ui/dist/"))
	mux.Handle("/", fileServer)

	mux.Handle("/library/v1/", catalog.MakeHandler(cs, httpLogger))
	http.Handle("/", accessControl(mux))
	http.Handle("/metrics", promhttp.Handler())


	errs := make(chan error, 2)
	go func() {
		logger.Log("transport", "http", "address", *httpAddr, "msg", "listening")
		errs <- http.ListenAndServe(*httpAddr, mux)
	}()
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	logger.Log("terminated", <-errs)
}

func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}

func envString(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}

