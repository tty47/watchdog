# WatchDog

WatchDog is a monitoring service that automatically detects Load Balancer resources in a Kubernetes cluster and exposes metrics related to these Load Balancers. The service uses OpenTelemetry to instrument the metrics and Prometheus to expose them.

---

## Overview

The Service WatchDog runs as a standalone service within the Kubernetes cluster. It periodically queries the Kubernetes API server to get a list of all services in the specified namespace. It then filters the list to include only services of type LoadBalancer. For each LoadBalancer service found, it retrieves the LoadBalancer IP and name and generates metrics with custom labels. These metrics are then exposed via a Prometheus endpoint, making them available for monitoring and visualization in Grafana or other monitoring tools.

---

## How It Works

WatchDog consists of two main components:

1. `main.go`: This is the main entry point of the Service WatchDog. It initializes the OpenTelemetry configuration, retrieves the list of Load Balancers using the `k8s.go` helper functions, and generates the metrics using the `WithMetricsLoadBalancer` function.

2. `k8s.go`: This file contains helper functions to interact with the Kubernetes API and retrieve the Load Balancers' information. The `GetNamespace` function gets the namespace where the service is deployed. The `ListServices` function retrieves the list of all services in the namespace, and the `GetLoadBalancers` function filters the list to include only LoadBalancer services and returns a list of `LoadBalancer` structs.

3. `metrics.go`: This file contains the OTEL functions, create metrics, expose, etc..

---

## Metrics

WatchDog generates the following custom metrics for each LoadBalancer:

- `load_balancer`: This metric represents the LoadBalancer resource and includes the following labels:
    - `service_name`: The service name. In this case, it is set to "watchdog."
    - `load_balancer_name`: The name of the LoadBalancer service.
    - `load_balancer_ip`: The IP address of the LoadBalancer.
    - `namespace`: The namespace in which the LoadBalancer is deployed.
    - `value`: The value of the metric. In this example, it is set to 1, but it can be customized to represent different load balancing states.

---

## Installation and Configuration

1. Clone this repository and navigate to the root folder.

2. Deploy the Service WatchDog to the Kubernetes cluster:

   ```bash
   kubectl apply -k deployment/overlays/local_dev
   ```

Access the Prometheus and Grafana dashboards to view and analyze the metrics exposed by the Service WatchDog.

---

## Monitoring and Visualization

WatchDog exposes the custom metrics through the Prometheus endpoint. You can use Grafana to connect to Prometheus and create custom dashboards to visualize the LoadBalancer metrics.

To access the Prometheus and Grafana dashboards and view the metrics, follow these steps:

1. Access the Prometheus dashboard:
  - Open a web browser and navigate to the Prometheus server's URL (e.g., `http://prometheus-server:9090`).
  - In the Prometheus web interface, you can explore and query the metrics collected by the Service WatchDog.

2. Access the Grafana dashboard:
  - Open a web browser and navigate to the Grafana server's URL (e.g., `http://grafana-server:3000`).
  - Log in to Grafana using your credentials.
  - Create a new dashboard or import an existing one to visualize the LoadBalancer metrics from Prometheus.
  - Use the `load_balancer` metric and its labels to filter and display the relevant information.

Customizing dashboards and setting up alerts in Grafana will help you monitor the performance and health of your LoadBalancer resources effectively.

---

## Customization

You can customize the Service WatchDog to suit your specific needs. For example, you can change the namespace where the service looks for LoadBalancer resources, adjust the metrics' names and labels, or modify the metrics' values based on your load balancing states.

---

Jose Ramon Mañes
