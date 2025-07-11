# Makefile for the 'ne' project to simplify Bazel commands.

# Phony targets are not associated with files, so they will always run.
.PHONY: all build test clean

# The default target when running 'make' without arguments.
all: build

# Build all targets using Bazel.
# This compiles the 'ne' and 'kvbuilder' binaries.
build:
	@echo "Building project with Bazel..."
	@bazel build //...

# Run all tests defined in the project using Bazel.
test:
	@echo "Running tests with Bazel..."
	@bazel test //...

# Clean all Bazel build artifacts, reset the cache, and remove old binaries.
clean:
	@echo "Cleaning Bazel artifacts and old binaries..."
	@bazel clean
	@rm -f ne kvbuilder
	@rm -f assets/ecdict.csv
