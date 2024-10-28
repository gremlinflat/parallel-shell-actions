# Variables
APP_NAME = parallel-shell-executor
SRC_DIR = src
DIST_DIR = dist
CONFIG_FILE = test/sample.json

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
	./$(DIST_DIR)/$(APP_NAME) -c $(CONFIG_FILE)