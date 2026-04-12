default: fmt lint install generate

# On Windows, GnuWin32 make needs Git's sh.exe to handle shell syntax (cd; go ...).
# If SHELL is not already pointing to a Unix shell, use Git for Windows.
ifeq ($(OS),Windows_NT)
  SHELL := C:/Program Files/Git/bin/sh.exe
endif

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

.PHONY: fmt lint test testacc build install generate
