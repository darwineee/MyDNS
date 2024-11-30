APP_NAME := mydns
OUTPUT_DIR := bin
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 windows/amd64
VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"

all: clean build-all run

.PHONY: build-all
build-all:
	@echo "Building $(APP_NAME) for multiple platforms..."
	@mkdir -p $(OUTPUT_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d'/' -f1) \
		GOARCH=$$(echo $$platform | cut -d'/' -f2) \
		go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(APP_NAME)-$$(echo $$platform | tr '/' '-')-$(VERSION) ./app/main.go || exit 1; \
	done
	@echo "Build completed."

.PHONY: build
build:
	@echo "Building $(APP_NAME) for the current platform..."
	@mkdir -p $(OUTPUT_DIR)
	go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(APP_NAME)-$(VERSION) ./app/main.go
	@echo "Build completed."

.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -f $(OUTPUT_DIR)/*
	@echo "Clean completed."

run: build
	@echo "Running $(APP_NAME)..."
	@$(OUTPUT_DIR)/$(APP_NAME)-$(VERSION)