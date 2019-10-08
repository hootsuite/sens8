IMAGE := hootsuite/sens8

.DEFAULT_GOAL := help
help: ## List targets & descriptions
	@cat Makefile* | grep -E '^[a-zA-Z_-]+:.*?## .*$$' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

all: test build build-image ## compile & build docker image

deps: ## install deps into vendor with golang dep
	go mod download

build: ## compile main go app
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o sens8 ./cmd/service

build-image: ## create docker image
	docker build -t $(IMAGE):latest .

test: ## run tests & coverage
	go test -v -cover $$(go list ./...)

docs: ## generate check command docs
	go run ./cmd/service/main.go -check-docs-md > check-docs.md
