package circuitbreaker_test

import (
	stdlog "log"
	"os"
	"testing"

	"github.com/afex/hystrix-go/hystrix"

	"github.com/go-kit/kit/circuitbreaker"
	kitlog "github.com/go-kit/kit/log"
)

func TestHystrix(t *testing.T) {
	logger := kitlog.NewLogfmtLogger(os.Stderr)
	stdlog.SetOutput(kitlog.NewStdlibAdapter(logger))

	const (
		commandName   = "my-endpoint"
		errorPercent  = 5
		maxConcurrent = 1000
	)
	hystrix.ConfigureCommand(commandName, hystrix.CommandConfig{
		ErrorPercentThreshold: errorPercent,
		MaxConcurrentRequests: maxConcurrent,
	})

	var (
		breaker          = circuitbreaker.Hystrix(commandName)
		primeWith        = hystrix.DefaultVolumeThreshold * 2
		shouldPass       = func(n int) bool { return (float64(n) / float64(primeWith+n)) <= (float64(errorPercent-1) / 100.0) }
		openCircuitError = hystrix.ErrCircuitOpen.Error()
	)
	testFailingEndpoint(t, breaker, primeWith, shouldPass, openCircuitError)
}
