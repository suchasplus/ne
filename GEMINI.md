# Gemini Project Context

This file provides context for the Gemini AI assistant to understand the `ne` project.

## Project Overview

`ne` is a command-line dictionary tool written in Go. It uses a local [BoltDB](https://github.com/etcd-io/bbolt) database for fast word lookups. The dictionary data is built from a source CSV file.

The core functionalities are:
1.  **Build**: Create or update the BoltDB database from a large CSV file (`ecdict.csv`).
2.  **Lookup**: Query the database for a word and display its definition, translation, phonetic transcription, and other details in a formatted table.
3.  **Compact**: Optimize the database file size and performance after large write operations (like building).

## Tech Stack

-   **Language**: Go
-   **Database**: BoltDB (via `go.etcd.io/bbolt`) for a local key-value store.
-   **CLI Framework**: `github.com/urfave/cli/v3`
-   **Logging**: `go.uber.org/zap`
-   **UI/Output Formatting**:
    -   `github.com/olekukonko/tablewriter` for creating formatted tables in the terminal.
    -   `github.com/charmbracelet/lipgloss` for styled/colored terminal output.

## Project Structure

```
.
├── assets/
│   ├── ecdict.csv.xz   # The compressed dictionary data source. MUST be decompressed before use.
│   └── README.md       # Instructions on how to decompress the data file.
├── cmd/
│   ├── ne/             # Main application entry point for the 'ne' CLI tool.
│   └── ...
└── internal/
    └── bbolthelper/    # A helper package to abstract BoltDB operations (open, close, get, put, import, compact).
```

-   **`cmd/ne/main.go`**: The main entry point for the application. It defines the CLI commands (`lookup`, `build`, `compact`) and their flags.
-   **`internal/bbolthelper/bbolthelper.go`**: Contains the `DBStore` struct and all logic for interacting with the BoltDB database. This includes serializing/deserializing data with `encoding/gob`.
-   **`assets/ecdict.csv.xz`**: The primary data source. It is compressed with `xz`.

## Development Workflow

### 1. Data Setup (One-time)

The dictionary data must be decompressed before it can be used to build the database.

```bash
# Navigate to the assets directory
cd assets

# Decompress the file
xz -d ecdict.csv.xz

# Navigate back to the project root
cd ..
```

### 2. Building the Application

The application can be built using the standard Go toolchain.

```bash
# Build the executable (output will be named 'ne')
go build -o ne ./cmd/ne
```

### 3. Building the Database

Once the application is built and the CSV is decompressed, create the BoltDB database.

```bash
# This command reads from assets/ecdict.csv and creates ecdict.bbolt
./ne build --csv assets/ecdict.csv
```

The `build` command will automatically run the `compact` operation upon completion.

### 4. Running the Application (Looking up words)

```bash
# Look up a word
./ne lookup <word>

# Example
./ne lookup hello
```

### 5. Running Tests

Standard Go test command.

```bash
go test ./...
```
