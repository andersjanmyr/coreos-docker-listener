REGISTRY = andersjanmyr
NAME = coreos-docker-listener
# VERSION will be sent as a parameter: make publish VERSION=1.1.1
VERSION = development
IMAGE = $(REGISTRY)/$(NAME)
IMAGE_WITH_VERSION = $(REGISTRY)/$(NAME):$(VERSION)
LATEST = $(REGISTRY)/$(NAME):latest

.PHONY: install, build, publish, image, run, run-local, bash, clean

install:
	go get github.com/coreos/go-etcd
	go get github.com/andersjanmyr/awsinfo


build:
	go build
	docker build --rm -t $(IMAGE_WITH_VERSION) .

publish: clean build
	docker tag -f $(IMAGE_WITH_VERSION) $(LATEST)
	docker push $(IMAGE)

image:
	@echo $(IMAGE_WITH_VERSION)

run:
	docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock $(IMAGE_WITH_VERSION)

run-local:
	docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock scratch

bash:
	docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock $(IMAGE_WITH_VERSION) sh

clean:
	go clean

