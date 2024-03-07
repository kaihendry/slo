# Service Level Objective (SLO) tutorial for Kubernetes 🐢

Goal is to show how to derive a Service Level Objective measure on a Kubernetes service "slo" using Prometheus metrics.

SLOs help teams create API **performance goals** and measure how well they are
meeting those goals.

The three pillars of observability are logs, metrics, and traces. The aim of this code is to show how **metrics** [are instumented](https://github.com/prometheus/client_golang), exported (kind: LoadBalancer), scraped (kind: PodMonitor) and queried (Prometheus Operator).

# Quickstart

1. `brew install colima hey helm` - install the tools
2. Give colima a bit more memory than defaults: `colima start --cpu 4 --memory 8`
3. Start Kubernetes cluster: `colima kubernetes start`
4. Install the [kube-prometheus-stack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm install tutorial prometheus-community/kube-prometheus-stack --set prometheus.prometheusSpec.podMonitorSelectorNilUsesHelmValues=false
```

5. `kubectl apply -f k8s/deploy.yaml` to deploy the "slo" service
6. Export prometheus service to localhost: `kubectl port-forward service/tutorial-kube-prometheus-s-prometheus 9090:9090`
7. Expose slo level 4 load balancer to localhost: `kubectl port-forward svc/slo-service 8080:8080`
8. `hey http://localhost:8080/` - generate 200 requests to the service
9. `hey "http://localhost:8080/?sleep=500"` - simulate 200 SLOW 🐢 responses frim the service

With a [Application Performance Index](https://en.wikipedia.org/wiki/Apdex) aka SLO query in Prometheus:

    sum(rate(request_duration_seconds_bucket{le="0.3"}[5m])) by (job)
    /
    sum(rate(request_duration_seconds_count[5m])) by (job)

[Documentation on these queries](https://prometheus.io/docs/practices/histograms/)

It should say 50% (0.5) of requests are under 300ms in the last 5 minutes. You
might **need to be patient** for the metrics to appear, as the /metrics
endpoints might not have been scraped by Prometheus yet.

![image](https://github.com/kaihendry/slo/assets/765871/6ebcb036-1da6-4489-ad66-207ae94a7208)

# Other resources

* https://github.com/kaihendry/pingprom - Prometheus quickstart with a black box ping exporter
* Old video explainer <https://www.youtube.com/watch?v=TNg3ga7s_MY&feature=youtu.be> that uses deprecated annotations

# Acknowledgements

shad and SuperQ on the Libera IRC #prometheus channel for helping.
