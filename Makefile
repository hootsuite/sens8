VERSION := $(shell cat VERSION)
IMAGE := hootsuite/sens8

.DEFAULT_GOAL := help
help: ## List targets & descriptions
	@cat Makefile* | grep -E '^[a-zA-Z_-]+:.*?## .*$$' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

all: deps build build-image ## compile & build docker image

deps: ## install deps into vendor with govendor
	go get github.com/kardianos/govendor
	govendor sync

build: ## compile main go app
	GOOS=linux GOARCH=amd64 go build -v -o sens8

build-image: ## create docker image
	docker build -t $(IMAGE):$(VERSION) .

push: ## docker push the image
	docker push $(IMAGE):$(VERSION)

tag-push-latest: ## tag the current version as latest and push
	docker tag $(IMAGE):$(VERSION) $(IMAGE):latest
	docker push $(IMAGE):latest

test: ## run tests & coverage
	go test -v -cover $$(go list ./... | grep -v /vendor/)

docs: ## generate check command docs
	go run sens8.go -check-commands-md > check-commands.md
