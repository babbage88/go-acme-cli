DOCKER_HUB:=ghcr.io/babbage88/goinfacli:
DOCKER_HUB_TEST:=jtrahan88/goinfacli-test:
ENV_FILE:=.env
MIG:=$(shell date '+%m%d%Y.%H%M%S')
SHELL := /bin/bash
VERBOSE ?= 0
ifeq ($(VERBOSE),1)
	V = -v
endif

build:
	go build $(V) -o goinfracli .

local-dev:
#	go build $(V) -o goinfracli .
	go build -o goinfracli -v

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

