IMAGE := hootsuite/sens8

.DEFAULT_GOAL := help
help: ## List targets & descriptions
	@cat Makefile* | grep -E '^[a-zA-Z_-]+:.*?## .*$$' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

all: deps test build build-image ## compile & build docker image

deps: ## install deps into vendor with golang dep
	go get github.com/golang/dep/cmd/dep
	dep ensure -v

build: ## compile main go app
	GOOS=linux GOARCH=amd64 go build -v -o sens8

build-image: ## create docker image
	docker build -t $(IMAGE):latest .

test: ## run tests & coverage
	go test -v -cover $$(go list ./... | grep -v /vendor/)

docs: ## generate check command docs
	go run sens8.go -check-docs-md > check-docs.md
