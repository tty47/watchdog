package main

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Get the meter from the global meter provider with the name "watchdog".
var meter = otel.GetMeterProvider().Meter("watchdog")

// LoadBalancer represents the information for a load balancer.
type LoadBalancer struct {
	ServiceName      string  // ServiceName Name of the service associated with the load balancer.
	LoadBalancerName string  // LoadBalancerName Name of the load balancer.
	LoadBalancerIP   string  // LoadBalancerIP IP address of the load balancer.
	Namespace        string  // Namespace where the service is deployed.
	Value            float64 // Value to be observed for the load balancer.
}

// WithMetricsLoadBalancer creates a callback function to observe metrics for multiple load balancers.
func WithMetricsLoadBalancer(loadBalancers []LoadBalancer) error {
	// Create a Float64ObservableGauge named "load_balancer" with a description for the metric.
	loadBalancersGauge, err := meter.Float64ObservableGauge(
		"load_balancer",
		metric.WithDescription("Service WatchDog - Load Balancers"),
	)
	if err != nil {
		log.Fatalf(err.Error())
		return err
	}

	// Define the callback function that will be called periodically to observe metrics.
	callback := func(ctx context.Context, observer metric.Observer) error {
		for _, lb := range loadBalancers {
			// Create labels with attributes for each load balancer.
			labels := metric.WithAttributes(
				attribute.String("service_name", lb.ServiceName),
				attribute.String("load_balancer_name", lb.LoadBalancerName),
				attribute.String("load_balancer_ip", lb.LoadBalancerIP),
				attribute.String("namespace", lb.Namespace),
			)
			// Observe the float64 value for the current load balancer with the associated labels.
			observer.ObserveFloat64(loadBalancersGauge, lb.Value, labels)
		}

		return nil
	}

	// Register the callback with the meter and the Float64ObservableGauge.
	_, err = meter.RegisterCallback(callback, loadBalancersGauge)
	return err
}
