all: build
.PHONY: all build

IMAGE ?= colonyos/pythonexecutor:v1.0.1

container:
	docker build -t $(IMAGE) .

push:
	docker push $(IMAGE)
