# Gemini Project Context

This file provides context for the Gemini AI assistant to understand the `ne` project.

## Project Overview

`ne` (stands for "玩转 English" - "Mastering English") is a command-line dictionary tool written in Go. It uses a local [BoltDB](https://github.com/etcd-io/bbolt) database for fast word lookups. The dictionary data is built from a source CSV file using a separate `kvbuilder` tool.

The core functionalities are split between two binaries:
1.  **`kvbuilder`**: A tool to create or update the BoltDB database from a large CSV file (`ecdict.csv`). It also compacts the database for optimal size.
2.  **`ne`**: The main CLI tool to query the database for a word and display its definition, translation, and other details in a formatted table.

## Tech Stack

-   **Language**: Go
-   **Database**: BoltDB (via `go.etcd.io/bbolt`) for a local key-value store.
-   **CLI Framework**: `github.com/urfave/cli/v3`
-   **Logging**: `go.uber.org/zap`
-   **UI/Output Formatting**:
    -   `github.com/charmbracelet/lipgloss/table` for creating formatted tables in the terminal.

## Project Structure

```
.
├── assets/
│   ├── ecdict.csv.xz   # The compressed dictionary data source. MUST be decompressed before use.
│   └── README.md       # Instructions on how to decompress the data file.
├── cmd/
│   ├── kvbuilder/      # The source for the 'kvbuilder' tool, which builds the database.
│   └── ne/             # The source for the 'ne' dictionary lookup tool.
└── internal/
    └── bbolthelper/    # A helper package to abstract BoltDB operations (open, close, get, put, import, compact).
```

-   **`cmd/kvbuilder/main.go`**: The entry point for the database builder tool.
-   **`cmd/ne/main.go`**: The main entry point for the dictionary lookup application. It takes the search term as a direct argument.
-   **`internal/bbolthelper/bbolthelper.go`**: Contains the `DBStore` struct and all logic for interacting with the BoltDB database.
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

### 2. Building the Applications

The project produces two executables.

```bash
# Build the database builder tool
go build -o kvbuilder ./cmd/kvbuilder

# Build the dictionary lookup tool
go build -o ne ./cmd/ne
```

### 3. Building the Database

Once the `kvbuilder` is built and the CSV is decompressed, create the BoltDB database.

**Usage:** `./kvbuilder [global options]`

**Example:**
```bash
# This command reads from assets/ecdict.csv and creates ecdict.bbolt
./kvbuilder --csv assets/ecdict.csv
```

**Global Options:**
-   `--csv FILE_PATH`, `-c FILE_PATH`: Load CSV from `FILE_PATH`.
-   `--dbpath string`, `-d string`: Path to bbolt DB.
-   `--bucket string`, `-b string`: Name of the bucket within the database.

The `kvbuilder` will automatically run a compaction operation upon completion.

### 4. Running the Application (Looking up words)

The `ne` tool takes the search term directly as an argument.

**Usage:** `./ne [global options] <term>`

**Example:**
```bash
# Look up a word using the 'ne' tool
./ne hello
```

**Global Options:**
-   `--verbose`, `-v`: Enable verbose logging output.
-   `--json`, `-j`, `-q`: Output result as JSON.
-   `--full`, `-f`: Show full map output in plain text (if not JSON).
-   `--dbpath string`, `-d string`: Path to the bbolt database file.
-   `--bucket string`, `-b string`: Name of the bucket within the database.

### 5. Running Tests

Standard Go test command.

```bash
go test ./...
```
