amd64:
	docker build -t sla --build-arg TARGET_ARCH=$@ .
	docker tag sla hendry/sla:$@
	docker push hendry/sla:$@

arm64:
	docker build -t sla --build-arg TARGET_ARCH=$@ .
	docker tag sla hendry/sla:$@
	docker push hendry/sla:$@

# https://github.com/golang/go/wiki/GoArm
arm:
	docker build --no-cache -t sla --build-arg TARGET_ARCH=$@ .
	docker tag sla hendry/sla:$@
	docker push hendry/sla:$@

manifest: amd64 arm64 arm
	docker manifest create --amend \
	  hendry/sla:latest \
	  hendry/sla:amd64 \
	  hendry/sla:arm

	docker manifest annotate hendry/sla:latest \
	  hendry/sla:amd64 --os linux --arch amd64

	docker manifest annotate hendry/sla:latest \
	  hendry/sla:arm --os linux --arch arm
	
	docker manifest inspect hendry/sla:latest
	docker manifest push hendry/sla:latest

run:
	docker run -it --env PORT=3000 -p 4000:3000 hendry/sla

deploy:
	docker push hendry/sla

orig:
	kubectl set image deployment/sla sla-sha256=asia.gcr.io/aliz-development/sla:orig

yawn:
	kubectl set image deployment/sla sla-sha256=asia.gcr.io/aliz-development/sla:yawn
