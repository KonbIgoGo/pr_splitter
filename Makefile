LOCAL_BIN := $(CURDIR)/bin
GOIMPORTS_BIN := $(LOCAL_BIN)/goimports
GOLANGCI_BIN := $(LOCAL_BIN)/golangci-lint
GO_TEST=$(LOCAL_BIN)/gotest
GO_TEST_ARGS=-race -v ./...

all: generate lint test

.PHONY: lint
lint:
	echo 'Running linter on files...'
	$(GOLANGCI_BIN) run \
	--config=.golangci.yaml \
	--sort-results \
	--max-issues-per-linter=0 \
	--max-same-issues=0


.PHONY: test
test:
	echo 'Running tests...'
	${GO_TEST} ${GO_TEST_ARGS}

bin-deps: .bin-deps

.bin-deps: export GOBIN := $(LOCAL_BIN)
.bin-deps: .create-bin 
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.5 && \
	GOBIN=$(LOCAL_BIN) go install github.com/rakyll/gotest@v0.0.6 && \
	go install golang.org/x/tools/cmd/goimports@v0.19.0 && \
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

.create-bin:
	rm -rf ./bin
	mkdir -p ./bin

generate: bin-deps .generate build
fast-generate: .generate

.generate:
	$(info Generating code...)

	rm -rf ./generated
	mkdir ./generated

	rm -rf ./docs/spec
	mkdir -p ./docs/spec

	rm -rf ~/.easyp/

	(PATH="$(PATH):$(LOCAL_BIN)")

	$(GOIMPORTS_BIN) -w .

	export PATH=$PATH:$(go env GOPATH)/bin
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml api/pr_splitter/openapi.yml
	go generate ./...

build:
	go mod tidy
	go build -o ./bin/pr_splitter ./cmd/pr_splitter/