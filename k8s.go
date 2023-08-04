package main

import (
	"context"
	"fmt"
	"log"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// GetNamespace retrieves the namespace where the service will be deployed
func GetNamespace() string {
	// Get the namespace from the environment variable "POD_NAMESPACE"
	// If the variable is not set, return the default namespace "default"
	namespace := os.Getenv("POD_NAMESPACE")
	if namespace == "" {
		return "default"
	}
	return namespace
}

// ListServices retrieves the list of services in a namespace
func ListServices(namespace string) (*corev1.ServiceList, error) {
	// Authentication in cluster - using Service Account, Role, RoleBinding
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// Create the Kubernetes clientSet
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// Get all services in the namespace
	services, err := clientSet.CoreV1().Services(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return services, nil
}

// GetLoadBalancers filters the list of services to include only Load Balancers and returns a list of them
func GetLoadBalancers(svc *corev1.ServiceList) []LoadBalancer {
	var loadBalancers []LoadBalancer

	for _, svc := range svc.Items {
		if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
			for _, ingress := range svc.Status.LoadBalancer.Ingress {
				fmt.Println("Updating metrics for:", svc.Name, ingress.IP)
				// Create a LoadBalancer struct and append it to the loadBalancers list
				loadBalancer := LoadBalancer{
					ServiceName:      "watchdog",
					LoadBalancerName: svc.Name,
					LoadBalancerIP:   ingress.IP,
					Namespace:        svc.Namespace,
					Value:            1, // Set the value of the metric here (e.g., 1)
				}
				loadBalancers = append(loadBalancers, loadBalancer)
			}
		}
	}

	return loadBalancers
}

// WatchServices watches for changes to the services in the specified namespace and updates the metrics accordingly
func WatchServices(namespace string) {
	// Authentication in cluster - using Service Account, Role, RoleBinding
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
		return
	}

	// Create the Kubernetes clientSet
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Create a service watcher
	watcher, err := clientSet.CoreV1().Services(namespace).Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
		return
	}

	// Watch for events on the watcher channel
	for event := range watcher.ResultChan() {
		if service, ok := event.Object.(*corev1.Service); ok {
			if service.Spec.Type == corev1.ServiceTypeLoadBalancer {
				// Get the list of Load Balancers from the single service
				loadBalancers := GetLoadBalancers(&corev1.ServiceList{Items: []corev1.Service{*service}})
				// Update the metrics with the Load Balancers
				WithMetricsLoadBalancer(loadBalancers)
			}
		}
	}
}
