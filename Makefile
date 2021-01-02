PWD := $(shell pwd)
GOPATH := $(shell go env GOPATH)
UIPATH := $(PWD)/ui


build:
	@echo "Building Library mangement to $(PWD)/main.go ..."
	@CGO_ENABLED=1 go build -o $(PWD) github.com/go-kit/kit


