NAME := hendry/sla
VERSION := $(shell git describe --tags --always)

manifest: amd64 arm
	docker manifest create --amend $(NAME):$(VERSION) $(NAME):amd64 $(NAME):arm
	docker manifest annotate $(NAME):$(VERSION) $(NAME):arm --os linux --arch arm
	docker manifest inspect $(NAME):$(VERSION)
	docker manifest push -p $(NAME):$(VERSION)

amd64:
	docker build -t $(NAME):$@ . --build-arg TARGET_ARCH=$@
	docker push $(NAME):$@

arm:
	docker build -t $(NAME):$@ . --build-arg TARGET_ARCH=$@
	docker push $(NAME):$@
