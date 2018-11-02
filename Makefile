.PHONY: default test build

OS := $(shell uname)
VERSION ?= 1.0.0

PKG_NAME = github.com/adhityaramadhanus/chronicle

# target #

default: test build

build: 
	@echo "Setup chronicle"
ifeq ($(OS),Linux)
	@echo "Build chronicle..."
	GOOS=linux  go build -ldflags "-s -w -X main.Version=$(VERSION)" -o chronicle cmd/server/main.go
endif
ifeq ($(OS) ,Darwin)
	@echo "Build chronicle..."
	GOOS=darwin go build -ldflags "-X main.Version=$(VERSION)" -o chronicle cmd/server/main.go
endif
	@echo "Succesfully Build for ${OS} version:= ${VERSION}"

# Test Packages

test:
	go test -v --cover ${PKG_NAME}
	# Integration Tests
	# go test -run Integration -v --cover ${PKG_NAME}/topic
	# go test -run Integration -v --cover ${PKG_NAME}/story
	GOCACHE=off go test -run Integration -v --cover ${PKG_NAME}/cmd/server
