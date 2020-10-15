package prometheus

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"democart/config"
	"democart/handler"
)

// NewHTTPServer constructs a new http.Server to listen for connections and
// serve responses for the prometheus metric collection service
func NewHTTPServer(configs *config.Configs) (*http.Server,
	handler.MiddlewareWrapper, error) {

	monitorHandler := http.NewServeMux()
	monitorHandler.Handle("/metrics", promhttp.Handler())

	s := &http.Server{
		Addr:    configs.MetricAddress,
		Handler: monitorHandler,
	}

	return s, basicMetricsMiddleware(configs.APISlug), nil
}

// basicMetricsMiddleware will return an http handler that can be used as
// middleware to wrap other servers and collect basic metrics on all requests
func basicMetricsMiddleware(service string) handler.MiddlewareWrapper {
	return func(h http.Handler) http.Handler {

		// TODO(sam): verify that these capture request path info
		h = promhttp.InstrumentHandlerInFlight(
			httpRequestInFlightGauge.With(
				prometheus.Labels{"service": service}),
			h)

		h = promhttp.InstrumentHandlerCounter(
			httpRequestCounter.MustCurryWith(
				prometheus.Labels{"service": service}),
			h)

		h = promhttp.InstrumentHandlerDuration(
			httpRequestDuration.MustCurryWith(
				prometheus.Labels{"service": service}),
			h)

		h = promhttp.InstrumentHandlerResponseSize(
			httpResponseSize.MustCurryWith(
				prometheus.Labels{"service": service}),
			h)

		return h
	}
}
