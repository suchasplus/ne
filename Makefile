# Makefile for the 'ne' project.

INSTALL_DIR := $(HOME)/.local/bin
CACHE_DIR   := $(HOME)/.cache/ne

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT)"

ECDICT_XZ   := assets/ecdict.csv.xz
ECDICT_CSV  := assets/ecdict.csv
CEDICT_GZ   := assets/cedict_1_0_ts_utf-8_mdbg.txt.gz
CEDICT_TXT  := assets/cedict_1_0_ts_utf-8_mdbg.txt

.PHONY: all build test clean install

# The default target when running 'make' without arguments.
all: build

# Full build: decompress assets -> compile binaries -> build bbolt databases.
build: $(ECDICT_CSV) $(CEDICT_TXT)
	@echo "==> Compiling binaries..."
	@go build -o kvbuilder ./cmd/kvbuilder
	@go build $(LDFLAGS) -o ne ./cmd/ne
	@echo "==> Building English dictionary database (ecdict.bbolt)..."
	@mkdir -p $(CACHE_DIR)
	@./kvbuilder --mode ecdict --csv $(ECDICT_CSV) --dbpath $(CACHE_DIR)/ecdict.bbolt
	@echo "==> Building Chinese dictionary database (cedict.bbolt)..."
	@./kvbuilder --mode cedict --csv $(CEDICT_TXT) --dbpath $(CACHE_DIR)/cedict.bbolt
	@echo "==> Build complete. Databases written to $(CACHE_DIR)/"

# Decompress English dictionary from .xz if not already done.
$(ECDICT_CSV): $(ECDICT_XZ)
	@echo "==> Decompressing $(ECDICT_XZ)..."
	@xz -dk $(ECDICT_XZ)

# Decompress Chinese dictionary from .gz if not already done.
$(CEDICT_TXT): $(CEDICT_GZ)
	@echo "==> Decompressing $(CEDICT_GZ)..."
	@gunzip -k $(CEDICT_GZ)

# Run all tests.
test:
	@echo "==> Running tests..."
	@go test ./...

# Install binaries to ~/.local/bin.
install: build
	@echo "==> Installing ne and kvbuilder to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	@cp ne $(INSTALL_DIR)/ne
	@cp kvbuilder $(INSTALL_DIR)/kvbuilder
	@echo "==> Installed. Make sure $(INSTALL_DIR) is in your PATH."

# Clean compiled binaries and decompressed asset files.
clean:
	@echo "==> Cleaning binaries and decompressed assets..."
	@rm -f ne kvbuilder
	@rm -f $(ECDICT_CSV) $(CEDICT_TXT)
