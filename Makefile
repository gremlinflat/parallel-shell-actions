# Variables
APP_NAME = parallel-shell-executor
SRC_DIR = src
DIST_DIR = dist
SAMPLE_INPUT = test/sample.json

# Default target
.PHONY: all
all: build

# Build the CLI application
.PHONY: build
build:
	@echo "Building the application..."
	go build -o $(DIST_DIR)/$(APP_NAME) $(SRC_DIR)/*.go

# Run the application with the default JSON configuration file
.PHONY: run
run: build
	@echo "Running the application..."
	@mkdir -p $(DIST_DIR)
	./$(DIST_DIR)/$(APP_NAME) -i $(SAMPLE_INPUT)


# Release
.PHONY: release
release: build
	@echo "Building the application for release..."
	@mkdir -p $(DIST_DIR)
	go build -o $(DIST_DIR)/$(APP_NAME) -ldflags "-s -w" $(SRC_DIR)/*.go
	@echo "Release build completed."
	@echo "You can find the release build in the '$(DIST_DIR)' directory."