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

// GetNamespace where the service will be deployed
func GetNamespace() string {
	namespace := os.Getenv("POD_NAMESPACE")
	if namespace == "" {
		return "default"
	}
	return namespace
}

// ListServices retrieves the load balancers in a namespace
func ListServices() (*corev1.ServiceList, error) {
	// get the namespace
	ns := GetNamespace()

	// authentication in cluster - using SA, Role, RoleBinding
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// creates the clientSet
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

// GetLoadBalancers filter by Load Balancers and return a list of them
func GetLoadBalancers(svc *corev1.ServiceList) []LoadBalancer {
	// get the namespace
	ns := GetNamespace()
	var loadBalancers []LoadBalancer

	// Filter only the services of type LoadBalancer
	for _, svc := range svc.Items {
		if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
			for _, ingress := range svc.Status.LoadBalancer.Ingress {
				fmt.Println("Updating metrics for:", svc.Name, ingress.IP)
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
