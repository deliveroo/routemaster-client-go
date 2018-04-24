GOPATH := $(shell go env GOPATH)
SRC_PARENT := ${GOPATH}/src/github.com/deliveroo
SRC := ${SRC_PARENT}/routemaster-client-go

.PHONY: all build install test

all: build test lint vet

$(SRC):
	@mkdir -p ${SRC_PARENT}
	@ln -sf $(shell pwd) ${SRC_PARENT}

link: $(SRC)

install:
	@cd ${SRC} && go install ./...

build: link install

lint: link
	@cd $(SRC) && golint

test: build
	@cd ${SRC} && go test -v

cover: build
	@cd ${SRC} && go test -cover

vet: build
	@cd ${SRC} && go vet
