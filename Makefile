# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
BINARY_NAME=refscaler
BINARY_UNIX=$(BINARY_NAME)_unix
BIN_DIR=bin
COVER_DIR=coverage
SRC_DIR=./...
VERSION=$(shell cat VERSION)

all: test build

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

build: $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/$(BINARY_NAME) -v

fmt:
	golangci-lint fmt $(SRC_DIR)

lint: fmt
	golangci-lint run

test:
	$(GOTEST) $(SRC_DIR)

test-cover: test
	mkdir -p $(COVER_DIR)
	$(GOTEST) -coverprofile=$(COVER_DIR)/coverage.out $(SRC_DIR)
	go tool cover -html=$(COVER_DIR)/coverage.out

clean:
	$(GOCLEAN)
	rm -rf $(BIN_DIR)

run: build
	./$(BIN_DIR)/$(BINARY_NAME)

tag: lint test
	git add .
	git commit -a -m "$(VERSION)"
	git push
	git tag -a $(VERSION) -m "$(VERSION)"
	git push origin $(VERSION)

.PHONY: \
	all \
	build \
	clean \
	fmt \
	lint \
	run \
	tag \
	test \
	test-cover
