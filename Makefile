DOCKER_HUB:=ghcr.io/babbage88/goinfacli:
BIN_NAME:=goinfracli
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
release:
	# 1. Ensure we're on the master branch
	@branch=$$(git rev-parse --abbrev-ref HEAD); \
	if [ "$$branch" != "master" ]; then \
	  echo "Error: You must be on the master branch. Current branch is '$$branch'."; \
	  exit 1; \
	fi; \
	echo "On master branch: $$branch"; \
	\
	# Ensure local master is up-to-date with the remote
	git fetch origin master; \
	UPSTREAM=origin/master; \
	LOCAL=$$(git rev-parse @); \
	REMOTE=$$(git rev-parse "$$UPSTREAM"); \
	BASE=$$(git merge-base @ "$$UPSTREAM"); \
	if [ "$$LOCAL" != "$$REMOTE" ]; then \
	  echo "Error: Your local master branch is not up-to-date with remote. Please pull the latest changes."; \
	  exit 1; \
	fi; \
	echo "Local master is up-to-date with remote."; \
	\
	# 2. Fetch all tags from remote
	git fetch --tags; \
	\
	# 3. Find the latest semver tag (vMAJOR.MINOR.PATCH)
	latest=$$(git tag -l "v[0-9]*.[0-9]*.[0-9]*" | sort -V | tail -n 1); \
	if [ -z "$$latest" ]; then \
	  echo "No semver tags found. Starting with v0.0.0"; \
	  latest="v0.0.0"; \
	fi; \
	echo "Latest tag: $$latest"; \
	\
	# Use bash to remove the leading "v" and split the version string by "."
	version=$${latest#v}; \
	IFS='.' read -r major minor patch <<< "$$version"; \
	echo "Current version: $$major.$$minor.$$patch"; \
	\
	# 4. Increment the chosen version type (default to patch)
	new_tag=$$(go run . utils version-bumper --latest-version $$latest --increment-type=$(VERSION)); \
	echo "Creating new tag: $$new_tag"; \
	\
	# 5. Create the new tag
	#git tag $$new_tag; \
	\
	# 6. Push the tag to remote
	#git push origin $$new_tag; \
	echo "Tag $$new_tag pushed to remote."


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

