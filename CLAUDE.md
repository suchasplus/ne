# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`ne` is a blazingly fast command-line dictionary tool written in Go that provides offline English dictionary lookups with fuzzy search capabilities. The project uses BoltDB for storage and implements intelligent word suggestions using Levenshtein distance.

## Architecture

The project consists of two main components:
- **kvbuilder**: Imports ECDICT CSV data (770k+ entries) into a BoltDB database
- **ne**: Command-line tool for dictionary lookups with fuzzy search

Key packages:
- `cmd/ne`: Main dictionary lookup CLI (uses urfave/cli/v3)
- `cmd/kvbuilder`: Database builder from CSV
- `internal/bbolthelper`: BoltDB wrapper with import/export functionality

## Build Commands

```bash
# Using Make (recommended)
make build          # Build all binaries using Bazel
make test           # Run all tests
make clean          # Clean build artifacts and old binaries

# Using Go directly
go build -o ne ./cmd/ne
go build -o kvbuilder ./cmd/kvbuilder
go test ./...

# Using Bazel directly
bazel build //...
bazel test //...
```

## Development Workflow

1. **Initial Setup**:
   ```bash
   cd assets && xz -d ecdict.csv.xz && cd ..
   ./kvbuilder --csv assets/ecdict.csv
   ```

2. **Testing Changes**:
   ```bash
   go test ./internal/bbolthelper -v  # Test specific package
   go test ./... -v                   # Test everything
   ```

3. **Running the Application**:
   ```bash
   ./ne <word>              # Basic lookup
   ./ne --json <word>       # JSON output
   ./ne --full <word>       # Show all fields
   ```

## Key Implementation Details

- **Fuzzy Search**: Uses Levenshtein distance with optimizations:
  - Length pruning (skip if length difference > 1)
  - Results sorted by word frequency (lower `frq` = higher frequency)
  - Returns top 3 suggestions from up to 10 candidates
  
- **Database Location**: `ecdict.bbolt` in current directory or `$HOME/.cache/ne/`

- **Normalization**: All keys are normalized to lowercase on import and lookup

- **Performance**: Smart linear scan for fuzzy search (future plans for Radix Tree implementation)

## Code Style

- Standard Go formatting and conventions
- Uses structured logging with zap
- Terminal UI styling with lipgloss
- Table output with tablewriter