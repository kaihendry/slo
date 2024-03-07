# Service Level Objective (SLO) tutorial for Kubernetes 🐢

Goal is to show how to derive a Service Level Objective measure on a Kubernetes cluster using Prometheus.

Assuming a fresh start:

1. `brew install colima hey helm` - install the tools
1. Give colima a bit more memory than defaults: `colima start --cpu 4 --memory 8`
2. Start Kubernetes cluster: `colima kubernetes start`
3. Install the [kube-prometheus-stack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm install tutorial prometheus-community/kube-prometheus-stack --set prometheus.prometheusSpec.podMonitorSelectorNilUsesHelmValues=false
```

4. `kubectl apply -f k8s/deploy.yaml` to deploy the "slo" service
5. Export prometheus service to localhost: `kubectl port-forward service/tutorial-kube-prometheus-s-prometheus 9090:9090`
6. Expose slo to localhost: `kubectl port-forward svc/slo-service 8080:8080`
7. `hey http://localhost:8080/` - generate 200 requests to the service
8. `hey http://localhost:8080/?sleep=500` - generate 200 SLOW 🐢 requests to the service

With a SLO query in Prometheus:

    sum(rate(request_duration_seconds_bucket{le="0.3"}[5m])) by (job)
    /
    sum(rate(request_duration_seconds_count[5m])) by (job)

It should say 50% of requests are under 300ms in the last 5 minutes. You might need to be patient for the metrics to appear.

![image](https://github.com/kaihendry/slo/assets/765871/6ebcb036-1da6-4489-ad66-207ae94a7208)

# Other resources

* https://github.com/kaihendry/pingprom - Prometheus quickstart with a black box ping exporter
* Old video explainer <https://www.youtube.com/watch?v=TNg3ga7s_MY&feature=youtu.be> that uses deprecated annotations

# Acknowledgements

shad and SuperQ on the Libera IRC #prometheus channel for helping.
