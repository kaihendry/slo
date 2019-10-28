build:
	docker build -t sla .
	docker tag sla hendry/sla

run:
	docker run -it --env PORT=3000 -p 4000:3000 sla

deploy:
	docker push hendry/sla

orig:
	kubectl set image deployment/sla sla-sha256=asia.gcr.io/aliz-development/sla:orig

yawn:
	kubectl set image deployment/sla sla-sha256=asia.gcr.io/aliz-development/sla:yawn
