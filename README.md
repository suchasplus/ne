# ne (玩转 English)

A blazingly fast command-line dictionary tool powered by Go and BoltDB.

`ne` (stands for "玩转 English" - "Mastering English") provides instant, offline access to a comprehensive English dictionary directly from your terminal.

## Features

-   **Offline First**: All lookups are performed locally. No internet connection required after initial setup.
-   **Extremely Fast**: Built on Go and using BoltDB, a high-performance key-value store, for near-instantaneous lookups.
-   **Simple & Clean UI**: Results are displayed in a clean, readable table format.
-   **Flexible Output**: Supports both human-readable tables and structured `JSON` output for scripting.
-   **Comprehensive Data**: Uses the extensive [ECDICT](https://github.com/skywind3000/ECDICT) dictionary data.

## Getting Started

### Prerequisites

-   **Go**: Version 1.24 or newer.
-   **xz**: A command-line tool for decompressing `.xz` files (e.g., `xz-utils` on Debian/Ubuntu, `xz` on macOS via Homebrew).

### Installation & Setup

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/suchasplus/ne.git
    cd ne
    ```

2.  **Decompress the Dictionary Data:**
    The dictionary source file comes compressed. You must decompress it before building the database.
    ```bash
    # Navigate to the assets directory
    cd assets

    # Decompress the file
    xz -d ecdict.csv.xz

    # Navigate back to the project root
    cd ..
    ```

3.  **Build the Tools:**
    This project uses two separate command-line tools: `kvbuilder` to build the database and `ne` to query it.
    ```bash
    # Build the database builder tool
    go build -o kvbuilder ./cmd/kvbuilder

    # Build the dictionary lookup tool
    go build -o ne ./cmd/ne
    ```
    You can move the `kvbuilder` and `ne` executables to a directory in your `$PATH` (e.g., `/usr/local/bin`) for easy access.

4.  **Build the Database:**
    Now, use the `kvbuilder` tool to create the local BoltDB database from the CSV file.
    ```bash
    # This command reads the CSV and creates the database file (ecdict.bbolt)
    ./kvbuilder --csv assets/ecdict.csv
    ```
    This process may take a minute. `kvbuilder` will create the `ecdict.bbolt` file in your current directory or in `$HOME/.cache/ne/` if it has permissions.

## Usage

### Looking Up a Word

To look up a word, simply pass it as an argument to the `ne` command.

**Syntax:**
```bash
./ne [options] <term>
```

**Example:**
```bash
./ne magnificent
```

**Options:**
-   `--json`, `-j`: Output the result in JSON format.
-   `--full`, `-f`: Show all available data fields for a term.
-   `--dbpath <path>`: Specify a custom path to the `ecdict.bbolt` database file.
-   `--verbose`, `-v`: Enable detailed logging.

### Rebuilding the Database

If you update the source `ecdict.csv` file, you can rebuild the database using `kvbuilder`.

**Syntax:**
```bash
./kvbuilder [options]
```
**Options:**
-   `--csv <path>`: Path to the source CSV file.
-   `--dbpath <path>`: Path where the output `ecdict.bbolt` database will be saved.
-   `--bucket <name>`: Specify a custom bucket name within the database.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
