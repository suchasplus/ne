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

---
---

## Architectural Design & Implementation Guide

This section provides a high-level design document for another Gemini agent to replicate the project. The focus is on the architecture, logic, and design principles, not on literal code.

### Phase 1: Project Scaffolding

> **Prompt:** "I'm building a Go-based CLI dictionary tool called `ne`. Initialize a Go module `github.com/suchasplus/ne`. Then, create the standard Go project layout: a `cmd` directory containing subdirectories for two separate tools, `ne` and `kvbuilder`. Also, create an `internal/bbolthelper` directory for shared database logic and an `assets` directory for data files. Finally, create a standard `.gitignore` file for Go projects that also ignores `*.bbolt` and `*.tmp` files."

### Phase 2: Core Database Helper (`internal/bbolthelper`)

This package is the heart of the project. Its purpose is to completely abstract away the complexities of BoltDB, providing a clean and simple API for the rest of the application.

> **Prompt:** "Create a database helper package in `internal/bbolthelper`. The goal is to wrap the `go.etcd.io/bbolt` library.
>
> **Design:**
> 1.  **`DBStore` Struct:** This will be the main struct, holding the `*bolt.DB` connection, a `*zap.Logger`, and configuration values like the DB path and bucket name.
> 2.  **`Config` Struct:** Create a `Config` struct to pass setup parameters (`DBPath`, `BucketName`, `ReadOnly`, `Logger`) to the constructor. This makes initialization clean and extensible.
> 3.  **Constructor `NewDBStore(Config)`:** This function should open the BoltDB file. A key feature is that if the database is opened in write mode, it must ensure the primary data bucket exists by using a `db.Update` transaction to call `CreateBucketIfNotExists`. It should also handle setting sensible defaults if parts of the `Config` are empty.
> 4.  **Serialization:** BoltDB only stores byte slices (`[]byte`). Since our dictionary entries are structured data, we need serialization. Implement two unexported functions: `serialize(map[string]string)` and `deserialize([]byte) (map[string]string, error)`. Use the standard `encoding/gob` package for this, as it's simple and built-in.
> 5.  **Public Methods:**
>     -   `Get(key string)`: This should use a read-only `db.View` transaction for high performance. It will fetch the byte slice, call `deserialize`, and return the `map[string]string`, a `bool` indicating if the key was found, and any error.
>     -   `Put(key string, value map[string]string)`: This should take a user-friendly map, call `serialize` internally, and use a `db.Update` transaction to save the data.
>     -   `ImportFromCSV(csvPath string, ...)`: This is a high-performance batch import function. The key design choice here is to perform the entire import within a **single `db.Update` transaction**. This is critical for performance, as committing after each record would be extremely slow. The logic should read the CSV header first, then iterate over each record, creating a map, serializing it, and putting it into the database.
>     -   `Compact()`: BoltDB doesn't reclaim space automatically. This method should implement the official compaction strategy: open a temporary DB file, use `tx.CopyFile` to efficiently write the data from the source DB to the temp DB, and then replace the original file with the temp one. It's important to note that the original `DBStore` instance becomes invalid after this operation.
>     -   `Close()`: A simple wrapper for `db.Close()`."

### Phase 3: The Database Builder (`cmd/kvbuilder`)

This is a dedicated, single-purpose tool for building the database.

> **Prompt:** "Now, create the `kvbuilder` CLI tool in `cmd/kvbuilder`. Use `urfave/cli/v3`.
>
> **Design:**
> 1.  **CLI Interface:** The tool should be simple. It needs flags for `--csv`, `--dbpath`, and `--bucket`. Provide sensible defaults, like searching for the CSV in `./assets` and the current directory, and defaulting the DB path to a user cache directory like `$HOME/.cache/ne/`.
> 2.  **Action Logic:**
>     -   The main action will instantiate the `bbolthelper.DBStore` in write mode.
>     -   It will then call `store.ImportFromCSV()` to perform the main task.
>     -   Crucially, after a successful import, it should immediately call `store.Compact()` to ensure the newly created database is optimized.
>     -   Provide clear logging to the user about the progress and completion."

### Phase 4: The Dictionary Tool (`cmd/ne`)

This is the user-facing tool. It should be fast, intuitive, and read-only.

> **Prompt:** "Finally, create the main `ne` tool in `cmd/ne`. Use `urfave/cli/v3` and `lipgloss/table`.
>
> **Design:**
> 1.  **CLI Interface:**
>     -   The primary interface is a direct argument: `ne <term>`. There is no `lookup` subcommand. The logic should be in the root command's `Action`.
>     -   Implement flags for controlling output and configuration: `--json` for machine-readable output, `--full` to show all dictionary fields, `--dbpath` to specify a non-standard DB location, and `--verbose` for debugging.
> 2.  **Action Logic:**
>     -   The action first validates that a search term was provided as an argument.
>     -   It then determines the database path, searching in common locations (current dir, `$PATH`, user cache) if the `--dbpath` flag isn't set.
>     -   It instantiates `bbolthelper.DBStore` in **read-only mode**. This is a key design choice for safety and performance.
>     -   It calls `store.Get()` with the search term.
>     -   It must handle the "not found" case gracefully, printing a user-friendly message.
>     -   If `--json` is specified, it marshals the result into a JSON object and prints it.
>     -   Otherwise, it formats the result into a clean, two-column table using `lipgloss/table`. By default, show only the most important fields (e.g., translation, definition). If `--full` is used, show all available fields."

### Phase 5: Final Touches

> **Prompt:** "Now that the code is logically complete, run `go mod tidy` to fetch all dependencies and populate `go.mod` and `go.sum`. Then, write a `README.md` that explains the project's purpose and provides clear setup and usage instructions for both `kvbuilder` and `ne`."