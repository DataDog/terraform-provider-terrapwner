MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
ROOT_DIR := $(dir $(MAKEFILE_PATH))

default: fmt lint install generate

build:
	go build -v ./...

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

thirdparty-licenses:
	@echo "Retrieving third-party licenses..."
	@dd-license-attribution https://github.com/DataDog/terraform-provider-terrapwner > $(ROOT_DIR)/LICENSE-3rdparty.csv
	@echo "Third-party licenses retrieved and saved to $(ROOT_DIR)/LICENSE-3rdparty.csv"

.PHONY: fmt lint test testacc build install generate thirdparty-licenses
