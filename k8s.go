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
func ListServices() (*corev1.ServiceList, error) {
	// Get the namespace
	ns := GetNamespace()

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
	services, err := clientSet.CoreV1().Services(ns).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return services, nil
}

// GetLoadBalancers filters the list of services to include only Load Balancers and returns a list of them
func GetLoadBalancers(svc *corev1.ServiceList) []LoadBalancer {
	// Get the namespace
	ns := GetNamespace()
	var loadBalancers []LoadBalancer

	// Filter only the services of type LoadBalancer
	for _, svc := range svc.Items {
		if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
			for _, ingress := range svc.Status.LoadBalancer.Ingress {
				fmt.Println("Updating metrics for:", svc.Name, ingress.IP)
				// Create a LoadBalancer struct and append it to the loadBalancers list
				loadBalancer := LoadBalancer{
					ServiceName:      "watchdog",
					LoadBalancerName: svc.Name,
					LoadBalancerIP:   ingress.IP,
					Namespace:        ns,
					Value:            1, // Set the value of the metric here (e.g., 1)
				}
				loadBalancers = append(loadBalancers, loadBalancer)
			}
		}
	}

	return loadBalancers
}
