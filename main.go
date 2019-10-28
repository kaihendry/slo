package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {

	inFlightGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "in_flight_requests",
		Help: "A gauge of requests currently being served by the wrapped handler.",
	})

	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_requests_total",
			Help: "A counter for requests to the wrapped handler.",
		}, []string{"code", "method"},
	)

	// duration is partitioned by the HTTP method and handler. It uses custom
	// buckets based on the expected request duration.
	duration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "A histogram of latencies for requests.",
			Buckets: []float64{.05, .1, .2, .3, .5},
		},
		[]string{"handler", "method"},
	)

	// Create the handlers that will be wrapped by the middleware.
	rootHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		fmt.Fprintln(w, fmt.Sprintf("Slept: %d ms..yawn", sleep))
	})

	// Register all of the metrics in the standard registry.
	prometheus.MustRegister(inFlightGauge, counter, duration)

	// Instrument the handlers with all the metrics, injecting the "handler" label by currying.
	rootChain := promhttp.InstrumentHandlerInFlight(inFlightGauge,
		promhttp.InstrumentHandlerDuration(duration.MustCurryWith(prometheus.Labels{"handler": "/"}),
			promhttp.InstrumentHandlerCounter(counter, rootHandler),
		),
	)
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/", rootChain)

	address := ":" + os.Getenv("PORT")
	log.Println("slake on " + address)
	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatal(err)
	}

}
