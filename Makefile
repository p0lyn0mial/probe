SHELL=/bin/bash
OUTPUT:=probe
BUILD_DIR?=_build

.PHONY: all clean build test fmt vet run

default: all

all: clean fmt vet test build

clean:
	rm -rf ./$(BUILD_DIR)/$(OUTPUT)
fmt:
	go fmt 

vet:
	go vet 

build: clean fmt vet
	go build -o ./$(BUILD_DIR)/$(OUTPUT) 

test:
	go test ./...

run:
	./$(BUILD_DIR)/$(OUTPUT)
