NAME := test
REGION := ap-southeast-1
ACCOUNT := 160071257600

VERSION  = $(shell git describe --tags --always --dirty)

all: $(VERSION)

login:
	aws ecr get-login-password --region ap-southeast-1 | docker login --username AWS --password-stdin 160071257600.dkr.ecr.ap-southeast-1.amazonaws.com

$(VERSION): options
	docker build -q -t $(NAME):$(VERSION) .
	docker tag $(NAME):$(VERSION) $(ACCOUNT).dkr.ecr.$(REGION).amazonaws.com/sre/test:$(VERSION)
	docker push $(ACCOUNT).dkr.ecr.$(REGION).amazonaws.com/$(NAME):$(VERSION)

options:
	@echo sla build options:
	@echo "VERSION   = ${VERSION}"

testci:
	gitlab-runner exec shell build-job


