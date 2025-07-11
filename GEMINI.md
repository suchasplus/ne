# Gemini Project Context

This file provides context for the Gemini AI assistant to understand the `ne` project.

## Project Overview

`ne` (stands for "玩转 English" - "Mastering English") is a command-line dictionary tool written in Go. It uses a local [BoltDB](https://github.com/etcd-io/bbolt) database for fast word lookups. The dictionary data is built from a source CSV file using a separate `kvbuilder` tool.

The core functionalities are split between two binaries:
1.  **`kvbuilder`**: A tool to create or update the BoltDB database from a large CSV file (`ecdict.csv`). It also compacts the database for optimal size.
2.  **`ne`**: The main CLI tool to query the database for a word and display its definition, translation, and other details in a formatted table. It includes a fuzzy search feature to correct common misspellings.

## Tech Stack

-   **Language**: Go
-   **Database**: BoltDB (via `go.etcd.io/bbolt`)
-   **CLI Framework**: `github.com/urfave/cli/v3`
-   **Fuzzy Search**: `github.com/agnivade/levenshtein`
-   **Logging**: `go.uber.org/zap`
-   **UI/Output Formatting**: `github.com/charmbracelet/lipgloss/table`

## Project Structure

```
.
├── assets/
│   ├── ecdict.csv.xz   # The compressed dictionary data source.
│   └── README.md       # Instructions on how to decompress the data file.
├── cmd/
│   ├── kvbuilder/      # Source for the 'kvbuilder' tool.
│   └── ne/             # Source for the 'ne' dictionary lookup tool.
└── internal/
    └── bbolthelper/    # Helper package for all BoltDB interactions.
```

## Architectural Design & Implementation Guide

This section provides a high-level design document for another Gemini agent to replicate the project.

### Phase 1: Project Scaffolding

> **Prompt:** "I'm building a Go-based CLI dictionary tool called `ne`. Initialize a Go module `github.com/suchasplus/ne`. Then, create the standard Go project layout: a `cmd` directory containing subdirectories for two separate tools, `ne` and `kvbuilder`. Also, create an `internal/bbolthelper` directory for shared database logic and an `assets` directory for data files. Finally, create a standard `.gitignore` file for Go projects that also ignores `*.bbolt` and `*.tmp` files."

### Phase 2: Core Database Helper (`internal/bbolthelper`)

This package abstracts all BoltDB complexities.

> **Prompt:** "Create a database helper package in `internal/bbolthelper`. The goal is to wrap `go.etcd.io/bbolt`.
>
> **Design:**
> 1.  **`DBStore` & `Config` Structs:** Create these to manage the DB connection and its configuration (`DBPath`, `BucketName`, `ReadOnly`, `Logger`).
> 2.  **Constructor `NewDBStore(Config)`:** This function should open the DB. If in write mode, it must ensure the main bucket exists using `CreateBucketIfNotExists`.
> 3.  **Serialization:** Implement unexported `serialize` and `deserialize` functions using `encoding/gob` to convert `map[string]string` to `[]byte` and back.
> 4.  **Public Methods:**
>     -   `Get(key string)`: Uses a read-only `db.View` transaction to fetch and deserialize a single record.
>     -   `FindSimilar(word string, maxDistance int)`: Implements fuzzy search. It iterates over all keys, calculating the Levenshtein distance. **Crucially, it must use performance optimizations**:
>         -   **Length Pruning**: Skip words where `abs(len(dbWord) - len(inputWord)) > maxDistance`.
>         -   **Dynamic Threshold Adjustment**: If a match with distance `d` is found, lower the `maxDistance` for subsequent searches to `d`.
>     -   `ImportFromCSV(...)`: A high-performance batch import function that runs inside a **single `db.Update` transaction**.
>     -   `Compact()`: Implements the official BoltDB compaction strategy (`tx.CopyFile`).
>     -   `Close()`: Wraps `db.Close()`."

### Phase 3: The Tools (`cmd/kvbuilder` and `cmd/ne`)

> **Prompt for `kvbuilder`:** "Create the `kvbuilder` CLI tool. Use `urfave/cli/v3`. It needs flags like `--csv` and `--dbpath`. Its main action should instantiate the `bbolthelper.DBStore`, call `ImportFromCSV()`, and then `Compact()`."
>
> **Prompt for `ne`:** "Create the main `ne` tool. Use `urfave/cli/v3` and `lipgloss/table`.
>
> **Design:**
> 1.  **Interface:** The main interface is a direct argument: `ne <term>`.
> 2.  **Flags:** Implement flags like `--json`, `--full`, and `--dbpath`.
> 3.  **Action Logic:**
>     -   Instantiate `DBStore` in **read-only mode**.
>     -   Call `store.Get()`.
>     -   **If not found**, automatically call `store.FindSimilar()`. If suggestions are returned, notify the user and call `store.Get()` on the best suggestion.
>     -   Format the final result as a table or JSON."

### Phase 4: Final Touches

> **Prompt:** "Now that the code is logically complete, run `go get` for the dependencies (`bbolt`, `urfave/cli`, `zap`, `lipgloss`, `levenshtein`). Then run `go mod tidy`. Finally, write a comprehensive `README.md` explaining the project's purpose, features (including fuzzy search), and clear setup/usage instructions."
