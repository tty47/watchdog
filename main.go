package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// var meter = otel.Meter("svcwatcher")
var meter = otel.GetMeterProvider().Meter("svcwatcher")

type LoadBalancer struct {
	ServiceName      string
	LoadBalancerName string
	LoadBalancerIP   string
	Value            float64
}

func WithMetrics(loadBalancers []LoadBalancer) error {
	loadBalancersGauge, err := meter.Float64ObservableGauge(
		"load_balancer",
		metric.WithDescription("Service Watch Dog - Load Balancers"),
	)
	if err != nil {
		log.Fatalf(err.Error())
		return err
	}

	callback := func(ctx context.Context, observer metric.Observer) error {
		for _, lb := range loadBalancers {
			labels := metric.WithAttributes(
				attribute.String("service_name", lb.ServiceName),
				attribute.String("load_balancer_name", lb.LoadBalancerName),
				attribute.String("load_balancer_ip", lb.LoadBalancerIP),
			)
			observer.ObserveFloat64(loadBalancersGauge, lb.Value, labels)
		}

		return nil
	}

	_, err = meter.RegisterCallback(callback, loadBalancersGauge)
	fmt.Println("-------------------------------------------")

	return err
}

func Run() {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf(err.Error())
	}
	// creates the clientSet
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Namespace where the service will be deployed
	namespace := os.Getenv("POD_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	// Get all services in the namespace
	services, err := clientSet.CoreV1().Services(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	var loadBalancers []LoadBalancer

	// Filter and log only the services of type LoadBalancer
	for _, svc := range services.Items {
		if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
			for _, ingress := range svc.Status.LoadBalancer.Ingress {
				fmt.Println("Updating metrics for:", svc.Name, ingress.IP)
				loadBalancer := LoadBalancer{
					ServiceName:      "svcwatcher",
					LoadBalancerName: svc.Name,
					LoadBalancerIP:   ingress.IP,
					Value:            1, // Set the value of the metric here (e.g., 1)
				}
				loadBalancers = append(loadBalancers, loadBalancer)
			}
		}
	}

	err = WithMetrics(loadBalancers)
	if err != nil {
		log.Printf("Failed to update metrics: %v", err)
	}
}

func main() {
	fmt.Println("Starting Service Watch Dog...")
	// Run the initial logic
	Run()

	exporter, err := prometheus.New()
	if err != nil {
		log.Fatal(err)
	}
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))

	otel.SetMeterProvider(provider)
	// Create a channel to receive termination signals
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	// Start an HTTP server in a goroutine
	go func() {
		log.Println("Starting HTTP server...")
		http.Handle("/metrics", promhttp.Handler())
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			Run()
		})
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for termination signal
	<-stopChan

	fmt.Println("Service Watch Dog has been stopped.")
}
