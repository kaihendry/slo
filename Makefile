build:
	ko build .

run:
	docker run --rm -p 8080:8080 -e PORT=8080 hendry/sla-d792c2e3a115ac1af16ceb6272431d48:latest

checkmetrics:
	curl -s localhost:8080/metrics | docker run --entrypoint=promtool prom/prometheus check metrics
