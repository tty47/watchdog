package main

import (
	"context"
	"fmt"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var meter = otel.GetMeterProvider().Meter("watchdog")

type LoadBalancer struct {
	ServiceName      string
	LoadBalancerName string
	LoadBalancerIP   string
	Value            float64
	Namespace        string
}

func WithMetricsLoadBalancer(loadBalancers []LoadBalancer) error {
	loadBalancersGauge, err := meter.Float64ObservableGauge(
		"load_balancer",
		metric.WithDescription("Service WatchDog - Load Balancers"),
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
				attribute.String("namespace", lb.Namespace),
			)
			observer.ObserveFloat64(loadBalancersGauge, lb.Value, labels)
		}

		return nil
	}

	_, err = meter.RegisterCallback(callback, loadBalancersGauge)
	fmt.Println("-------------------------------------------")

	return err
}
