package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"time"

	_ "net/http/pprof"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	Version   string
	Revision  string
	Branch    string
	BuildUser string
	BuildDate string
	GoVersion = runtime.Version()
)

// Map provides the iterable version information.
var Map = map[string]string{
	"version":   Version,
	"revision":  Revision,
	"branch":    Branch,
	"buildUser": BuildUser,
	"buildDate": BuildDate,
	"goVersion": GoVersion,
}

func init() {
	buildInfo := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sla_build_info",
			Help: "A metric with a constant '1' value labeled by version, revision, branch, and goversion from which sla was built.",
		},
		[]string{"version", "revision", "branch", "goversion"},
	)
	buildInfo.WithLabelValues(Version, Revision, Branch, GoVersion).Set(1)
	prometheus.MustRegister(buildInfo)
}

// Create the handlers that will be wrapped by the middleware.
func root(w http.ResponseWriter, r *http.Request) {
	sleep, err := strconv.Atoi(r.URL.Query().Get("sleep"))
	if err == nil {
		log.Println(fmt.Sprintf("Sleeping: %d milliseconds", sleep))
		time.Sleep(time.Duration(sleep) * time.Millisecond)
	}
	code, err := strconv.Atoi(r.URL.Query().Get("code"))
	if err == nil {
		if code >= 200 {
			log.Println(fmt.Sprintf("Code: %d", code))
			w.WriteHeader(code)
		}
	}
	fmt.Fprintln(w, fmt.Sprintf("Slept: %d ms", sleep))
}

func main() {

	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_duration_seconds_count",
			Help: "A counter for requests to the wrapped handler.",
		}, []string{"code", "method"},
	)

	// duration is partitioned by the HTTP method and handler. It uses custom
	// buckets based on the expected request duration.
	//
	// sum(rate(http_request_duration_seconds_bucket{le="0.3"}[5m])) by (job)
	// /
	// sum(rate(http_request_duration_seconds_count[5m])) by (job)

	duration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "A histogram of latencies for requests.",
			// 50ms, 100ms, 200ms, 300ms, 500ms
			Buckets: []float64{.05, .1, .2, .3, .5},
		},
		[]string{"handler", "method"},
	)

	// Register all of the metrics in the standard registry.
	prometheus.MustRegister(counter, duration)

	// Pprof server.
	// https://mmcloughlin.com/posts/your-pprof-is-showing
	go func() {
		log.Fatal(http.ListenAndServe(":8081", nil))
	}()

	// Application server.
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	// Injecting the "handler" label by currying.
	promhttp.InstrumentHandlerDuration(duration.MustCurryWith(prometheus.Labels{"handler": "/"}),
		promhttp.InstrumentHandlerCounter(counter, http.HandlerFunc(root)),
	)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
