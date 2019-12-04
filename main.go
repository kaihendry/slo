package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	_ "net/http/pprof"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	Version   string
	Branch    string
	BuildUser string
	BuildHost string
	GoVersion = runtime.Version()

	inFlightGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "in_flight_requests",
		Help: "A guage of requests currently being served by the wrapped handler",
	})

	counter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "requests_total",
		Help: "A counter for requests to the wrapped handler",
	},
		[]string{"code", "method"})

	duration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "request_duration_seconds",
			Help: "A histogram of latencies for requests.",
			// 50ms, 100ms, 200ms, 300ms, 500ms
			Buckets: []float64{.05, .1, .2, .3, .5},
		},
		[]string{"handler", "code", "method"},
	)

	buildInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sla_build_info",
			Help: "A metric with a constant '1' value labeled by attributes from which sla was built.",
		},
		[]string{"version", "branch", "buildUser", "goversion"},
	)
)

func init() {
	prometheus.MustRegister(inFlightGauge)
	prometheus.MustRegister(counter)
	prometheus.MustRegister(duration)
	prometheus.MustRegister(buildInfo)
	buildInfo.WithLabelValues(Version, Branch, BuildUser, GoVersion).Set(1)
}

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

	// duration is partitioned by the HTTP method and handler. It uses custom
	// buckets based on the expected request duration.
	// https://prometheus.io/docs/practices/histograms/
	// sum(rate(request_duration_seconds_bucket{le="0.3"}[5m])) by (job)
	// /
	// sum(rate(request_duration_seconds_count[5m])) by (job)

	// Register all of the metrics in the standard registry.

	// Pprof server.
	// https://mmcloughlin.com/posts/your-pprof-is-showing
	// go func() {
	// 	log.Fatal(http.ListenAndServe(":8081", nil))
	// }()

	// Application server.
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	rootChain := promhttp.InstrumentHandlerInFlight(
		inFlightGauge,
		promhttp.InstrumentHandlerDuration(
			duration.MustCurryWith(prometheus.Labels{"handler": "root"}),
			promhttp.InstrumentHandlerCounter(counter, http.HandlerFunc(root))))

	mux.Handle("/", rootChain)

	if err := http.ListenAndServe(":"+os.Getenv("PORT"), mux); err != nil {
		log.Fatal(err)
	}
}
