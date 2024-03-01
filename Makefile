build:
	docker build -t hendry/slo:latest .
	
# publish to https://hub.docker.com/repository/docker/hendry/slo
push:
	docker push hendry/slo:latest

run:
	docker run --rm -p 8080:8080 -e PORT=8080 hendry/slo:latest

checkmetrics:
	curl -s localhost:8080/metrics | docker run --entrypoint=promtool prom/prometheus check metrics
