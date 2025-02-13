DOCKER_HUB:=ghcr.io/babbage88/goinfacli:
BIN_NAME:=goinfracli
INSTALL_PATH:=$${GOPATH}bin
ENV_FILE:=.env
MIG:=$(shell date '+%m%d%Y.%H%M%S')
SHELL := /bin/bash
VERBOSE ?= 1
ifeq ($(VERBOSE),1)
	V = -v
endif

build:
	go build $(V) -o $(BIN_NAME) .

build-quiet:
	go build -o goinfracli

install: build
	echo "Install Path $(INSTALL_PATH)"
	mv $(BIN_NAME) $(INSTALL_PATH)

check-builder:
	@if ! docker buildx inspect goinfaclibuilder > /dev/null 2>&1; then \
		echo "Builder goinfaclibuilder does not exist. Creating..."; \
		docker buildx create --name goinfaclibuilder --bootstrap; \
	fi

create-builder: check-builder

buildandpush: check-builder
	docker buildx use goinfaclibuilder
	docker buildx build --platform linux/amd64,linux/arm64 -t $(DOCKER_HUB)$(tag) . --push

deploydev: buildandpushdev
	kubectl apply -f deployment/kubernetes/infra-goinfacli.yaml
	kubectl rollout restart deployment infra-goinfacli

