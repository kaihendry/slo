package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	Version   string
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

	duration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "request_duration_seconds",
			Help: "A histogram of latencies for requests.",
			// 50ms, 100ms, 200ms, 300ms, 500ms
			Buckets: []float64{.05, .1, .2, .3, .5},
		},
		[]string{"handler", "code", "method"},
	)

	buildInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "slo_build_info",
			Help: "A metric with a constant '1' value labeled by attributes from which slo was built.",
		},
		[]string{"version", "goversion"},
	)
)

func requestLog(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
		slog.Debug("request", "method", r.Method, "url", r.URL.Path)
	})
}

func root(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	sleep, err := strconv.Atoi(r.URL.Query().Get("sleep"))
	if err == nil {
		slog.Info("sleeping", "ms", sleep)
		time.Sleep(time.Duration(sleep) * time.Millisecond)
	}
	code, err := strconv.Atoi(r.URL.Query().Get("code"))
	if err == nil {
		if code >= 200 {
			slog.Warn("overriding status code", "code", code)
			w.WriteHeader(code)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Version", Version)

	response := map[string]interface{}{
		"url":     "https://github.com/kaihendry/slo",
		"elapsed": time.Since(start).Milliseconds(),
		"slept":   sleep,
		"version": Version,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		slog.Error("error encoding response", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	slog.SetDefault(getLogger(os.Getenv("LOGLEVEL")))
	Version, _ = GitCommit()

	buildInfo.WithLabelValues(Version, GoVersion).Set(1)

	// duration is partitioned by the HTTP method and handler. It uses custom
	// buckets based on the expected request duration.
	// https://prometheus.io/docs/practices/histograms/
	// sum(rate(request_duration_seconds_bucket{le="0.3"}[5m])) by (job)
	// /
	// sum(rate(request_duration_seconds_count[5m])) by (job)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	rootChain := promhttp.InstrumentHandlerInFlight(
		inFlightGauge,
		promhttp.InstrumentHandlerDuration(
			duration.MustCurryWith(prometheus.Labels{"handler": "root"}),
			promhttp.InstrumentHandlerCounter(counter, http.HandlerFunc(root))))

	mux.Handle("/", rootChain)

	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}
	if _, err := strconv.Atoi(port); err != nil {
		slog.Error("invalid port", "port", port)
		os.Exit(1)
	}

	slog.Info("starting slo", "port", port, "Version", Version, "GoVersion", GoVersion)

	if err := http.ListenAndServe(":"+port, requestLog(mux)); err != nil {
		slog.Error("error listening", err)
	}
}

func getLogger(logLevel string) *slog.Logger {
	levelVar := slog.LevelVar{}

	if logLevel != "" {
		if err := levelVar.UnmarshalText([]byte(logLevel)); err != nil {
			panic(fmt.Sprintf("Invalid log level %s: %v", logLevel, err))
		}
	}

	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: levelVar.Level(),
	}))
}

func GitCommit() (commit string, dirty bool) {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "", false
	}
	for _, setting := range bi.Settings {
		switch setting.Key {
		case "vcs.modified":
			dirty = setting.Value == "true"
		case "vcs.revision":
			commit = setting.Value
		}
	}
	return
}
