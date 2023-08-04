package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// GetHttpPort GetPort retrieves the namespace where the service will be deployed
func GetHttpPort() string {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		log.Println("Using the default port: 8080")
		return "8080"
	}

	// Ensure that the provided port is a valid numeric value
	_, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("Invalid HTTP_PORT value: %v. Using default port 8080")
		return "8080"
	}

	return port
}

// Run generates the services metrics
func Run() {
	// Get the namespace
	ns := GetNamespace()

	// Get list of LBs
	svc, err := ListServices(ns)
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

	// Start watching for changes to the services in a separate goroutine
	go WatchServices(ns)
}

// InitConfig initializes the configs Prometheus - OTEL
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
	// Get http port
	httpPort := GetHttpPort()

	fmt.Println("[INFO] Starting Service WatchDog...")

	// Initialize the config
	InitConfig()
	// Generate initial metrics
	Run()

	// Create a channel to receive termination signals
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	// Create a new router using gorilla/mux
	r := mux.NewRouter()

	// Add the logging middleware to all routes
	r.Use(LogRequest)

	// Define the main handler function
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Your main logic here
		Run()
	})
	// Register the prometheus endpoint
	r.Handle("/metrics", promhttp.Handler())

	// Start an HTTP server in a goroutine
	go func() {
		log.Println("[INFO] Starting HTTP server listening on port:", httpPort)
		if err := http.ListenAndServe(":"+httpPort, r); err != nil {
			log.Fatalf("[ERROR] HTTP server error: %v", err)
		}
	}()

	// Wait for termination signal
	<-stopChan

	fmt.Println("[INFO] Service Watch Dog has been stopped.")
}

// LogRequest is a middleware function that logs the incoming request.
func LogRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, " ", r.URL.Path)
		handler.ServeHTTP(w, r)
	})
}
