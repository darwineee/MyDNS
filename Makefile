APP_NAME := mydns
OUTPUT_DIR := bin

all: clean build run

.PHONY: build
build:
	@echo "Building $(APP_NAME)..."
	mkdir -p $(OUTPUT_DIR)
	go build -o $(OUTPUT_DIR)/$(APP_NAME) ./app/main.go

.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -f $(OUTPUT_DIR)/*

run:
	@echo "Running $(APP_NAME)..."
	$(OUTPUT_DIR)/$(APP_NAME)