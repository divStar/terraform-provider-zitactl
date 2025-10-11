ARTIFACT_NAME := $(shell grep "^module " go.mod | awk '{print $$2}' | xargs basename)

default: fmt lint install generate

build:
	go build -v ./...

artifact:
	go build -gcflags="all=-N -l" -o $(ARTIFACT_NAME)

install: build
	go install -v ./...

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./...

.PHONY: fmt lint test testacc build install generate
