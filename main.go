package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// HTTP listening port
const port = "8080"

// Run generate the services metrics
func Run() {
	// Get list of LBs
	svc, err := ListServices()
	if err != nil {
		log.Printf("Failed to retrieve the LoadBalancers: %v", err)
	}

	// Get the list of the LBs
	loadBalancers := GetLoadBalancers(svc)

	// Generate the metrics with the LBs
	err = WithMetricsLoadBalancer(loadBalancers)
	if err != nil {
		log.Printf("Failed to update metrics: %v", err)
	}
}

// InitConfig initialize the configs Prometheus - OTEL
func InitConfig() {
	// Initialize the Prometheus exporter
	exporter, err := prometheus.New()
	if err != nil {
		log.Fatal(err)
	}
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))

	otel.SetMeterProvider(provider)
}

func main() {
	fmt.Println("[INFO] Starting Service WatchDog...")

	// Initialize the config
	InitConfig()
	// Generate initial metrics
	Run()

	// Create a channel to receive termination signals
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	// Start an HTTP server in a goroutine
	go func() {
		log.Println("[INFO] Starting HTTP server listening on port:", port)
		// prometheus endpoint
		http.Handle("/metrics", promhttp.Handler())
		// Run function
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			Run()
		})
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("[ERROR] HTTP server error: %v", err)
		}
	}()

	// Wait for termination signal
	<-stopChan

	fmt.Println("[INFO] Service Watch Dog has been stopped.")
}
