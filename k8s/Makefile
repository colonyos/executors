all: build
.PHONY: all build

IMAGE ?= colonyos/k8sexecutor:v0.0.4

VERSION := $(shell git rev-parse --short HEAD)
BUILDTIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

GOLDFLAGS += -X 'main.BuildVersion=$(VERSION)'
GOLDFLAGS += -X 'main.BuildTime=$(BUILDTIME)'

build:
	@CGO_ENABLED=0 go build -ldflags="-s -w $(GOLDFLAGS)" -o ./bin/k8s_executor ./cmd/main.go

container:
	docker build -t $(IMAGE) .

push:
	docker push $(IMAGE)
