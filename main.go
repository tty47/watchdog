package main

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var meter = otel.Meter("svcwatcher")

func WithMetrics(loadBalancer string) error {
	loadBalancersGauge, err := meter.Float64ObservableGauge(
		"build_info",
		metric.WithDescription("Celestia Node build information"),
	)
	if err != nil {
		return err
	}

	callback := func(ctx context.Context, observer metric.Observer) error {
		// Observe build info with labels
		labels := metric.WithAttributes(
			attribute.String("load_balancer", loadBalancer),
		)

		fmt.Println(labels)
		observer.ObserveFloat64(loadBalancersGauge, 1, labels)

		return nil
	}

	_, err = meter.RegisterCallback(callback, loadBalancersGauge)

	return err
}

func Run() {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Namespace where the service will be deployed
	namespace := os.Getenv("POD_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	// Get all services in the namespace
	services, err := clientset.CoreV1().Services(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	// Filter and log only the services of type LoadBalancer
	for _, svc := range services.Items {
		if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
			log.Println("----------------------")
			log.Println("Name:", svc.Name)
			log.Println("Namespace:", svc.Namespace)
			log.Println("ClusterIP:", svc.Spec.ClusterIP)
			for _, ingress := range svc.Status.LoadBalancer.Ingress {
				log.Println("Public IP:", ingress.IP)
				// Register metrics with load balancer attribute
				err := WithMetrics(ingress.IP)
				if err != nil {
					log.Printf("Failed to register metrics for load balancer %s: %v", ingress.IP, err)
				}
			}
			// Add additional information you want to log
		}
	}

	// Expose the services via OTEL (assuming you have the necessary OTEL components configured)
	// Your OTEL configuration and logic here
}

func main() {
	fmt.Println("Starting Service Watch Dog...")

	// Run the initial logic
	Run()

	// Create a channel to receive termination signals
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	// Start an HTTP server in a goroutine
	go func() {
		log.Println("Starting HTTP server...")
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Service Watch Dog is running!")
		})
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for termination signal
	<-stopChan

	fmt.Println("Service Watch Dog has been stopped.")
}
