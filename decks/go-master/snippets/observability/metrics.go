package observability

import (
	"log/slog"
	"time"

	"github.com/uber-go/tally/v4"
	"github.com/uber-go/tally/v4/prometheus"
	"go.temporal.io/sdk/client"
	sdktally "go.temporal.io/sdk/contrib/tally"
	"go.temporal.io/sdk/log"
)

// slide: 25-observability.md "Integration: SDK metrics in Go"
func Dial(hostPort string) (client.Client, error) {
	reporter, err := prometheus.Configuration{
		ListenAddress: "0.0.0.0:9000",
		TimerType:     "histogram",
	}.NewReporter(prometheus.ConfigurationOptions{})
	if err != nil {
		return nil, err
	}
	scope, _ := tally.NewRootScope(tally.ScopeOptions{
		CachedReporter:  reporter,
		Separator:       prometheus.DefaultSeparator,
		SanitizeOptions: &sdktally.PrometheusSanitizeOptions,
	}, time.Second)

	return client.Dial(client.Options{
		HostPort:       hostPort,
		Logger:         log.NewStructuredLogger(slog.Default()),
		MetricsHandler: sdktally.NewMetricsHandler(sdktally.NewPrometheusNamingScope(scope)),
	})
}

// end slide
