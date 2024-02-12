# Directories
TMP_DIR := tmp
BIN_DIR := bin
CMD_DIR := cmd

# Commands
GENERATE_CMD := templ generate
BUILD_CMD := go build
FORMAT_CMD := gofmt
GO_MAIN := $(CMD_DIR)/main.go

# Determine operating system
ifeq ($(OS),Windows_NT)
	OS_NAME := windows
	EXEC_EXT := .exe
	RM := powershell -Command "Remove-Item -Force -Recurse"
else
	OS_NAME := $(shell uname -s | tr '[:upper:]' '[:lower:]')
	EXEC_EXT :=
	RM := rm -rf
endif

# Targets
.PHONY: prebuild build build-optimized clean all

# prebuild: Generate and build to temporary directory
prebuild: 
	$(GENERATE_CMD)
	$(BUILD_CMD) -o $(TMP_DIR)/main$(EXEC_EXT) $(GO_MAIN)

# build: Build the application to the bin directory
build: 
	$(BUILD_CMD) -o $(BIN_DIR)/app$(EXEC_EXT) $(GO_MAIN)

# build-optimized: Build optimized app executable with netgo tags, stripped debug info, suppressed linker warnings
build-optimized: 
	$(BUILD_CMD) -tags netgo -ldflags '-s -w' -o app $(GO_MAIN)

# clean: Remove temporary and bin directories
clean:
	$(RM) $(TMP_DIR) $(BIN_DIR)

fmt:
	$(FORMAT_CMD) cmd handlers internal models services templates

# all: Run prebuild and build
all: prebuild build

# # curl
# ADDR := http://localhost:1234
# CURL_HEADER := curl -I $(ADDR)
# curl-header:
# 	$(CURL_HEADER)

