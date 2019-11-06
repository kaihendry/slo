NAME := hendry/sla

VERSION  = $(shell git describe --tags --always --dirty)
BRANCH   = $(shell git rev-parse --abbrev-ref HEAD)
DATE     = $(shell date +'%FT%T%z')
HOSTNAME = $(shell hostname -s)

all: $(VERSION) latest

$(VERSION) latest: options amd64 arm
	docker manifest create --amend $(NAME):$@ $(NAME):amd64 $(NAME):arm
	docker manifest annotate $(NAME):$@ $(NAME):arm --os linux --arch arm
	docker manifest inspect $(NAME):$@
	docker manifest push -p $(NAME):$@

arm amd64:
	docker build -q -t $(NAME):$@ . \
		--build-arg TARGET_ARCH=$@ \
		--build-arg VERSION=$(VERSION) \
		--build-arg BRANCH=$(BRANCH) \
		--build-arg USER=$(USER) \
		--build-arg BUILDDATE=$(DATE) \
		--build-arg HOST=$(HOSTNAME)
	docker push $(NAME):$@

options:
	@echo sla build options:
	@echo "VERSION   = ${VERSION}"
	@echo "BRANCH    = ${BRANCH}"
	@echo "DATE      = ${DATE}"
	@echo "HOSTNAME  = ${HOSTNAME}"

run:
	docker run -p 8080:8080 -p 8081:8081 $(NAME)
