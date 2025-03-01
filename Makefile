DOCKER_HUB:=ghcr.io/babbage88/goinfacli:
BIN_NAME:=goinfracli
VERSION_TYPE:=patch
INSTALL_PATH:=$${GOPATH}/bin
ENV_FILE:=.env
MIG:=$(shell date '+%m%d%Y.%H%M%S')
SHELL := /bin/bash
VERBOSE ?= 1
ifeq ($(VERBOSE),1)
	V = -v
endif

sqlc-and-migrations:
	source config_goose.sh
	goose down -v
	goose up -v
	sqlc generate

build:
	go build $(V) -o $(BIN_NAME) .

build-quiet:
	go build -o goinfracli

install: build
	echo "Install Path $(INSTALL_PATH)"
	mv $(BIN_NAME) $(INSTALL_PATH)

# Add this target to the end of your Makefile

# Usage: make release [VERSION=major|minor|patch]
fetch-tags:
	@git fetch --tags
release: fetch-tags
	$(eval LATEST_TAG := $(shell git tag -l "v[0-9]*.[0-9]*.[0-9]*" | sort -V | tail -n 1))
	@branch=$$(git rev-parse --abbrev-ref HEAD);
	echo "On master branch: $$branch";
	git fetch origin master;
	UPSTREAM=origin/master;
	LOCAL=$$(git rev-parse @)
	REMOTE=$$(git rev-parse "$$UPSTREAM");
	BASE=$$(git merge-base @ "$$UPSTREAM");
	if [ "$$LOCAL" != "$$REMOTE" ]; then echo "Error: Your local master branch is not up-to-date with remote. Please pull the latest changes."
	  exit 1; 
	fi; 
	echo "Local master is up-to-date with remote."
	echo "Latest tag: $(LATEST_TAG)"
	new_tag=$$(go run . utils version-bumper --latest-version $(LATEST_TAG)--increment-type=$(VERSION_TYPE)); \
	echo "Creating new tag: $$new_tag"
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

