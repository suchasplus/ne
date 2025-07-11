# ne (玩转 ENglish)

A blazingly fast command-line dictionary tool powered by Go and BoltDB.

`ne` (stands for "玩转 ENglish" - "Mastering ENglish") provides instant, offline access to a comprehensive English dictionary directly from your terminal.

## Features

-   **Offline First**: All lookups are performed locally. No internet connection required after initial setup.
-   **Extremely Fast**: Built on Go and using BoltDB, a high-performance key-value store, for near-instantaneous lookups.
-   **Simple & Clean UI**: Results are displayed in a clean, readable table format.
-   **Flexible Output**: Supports both human-readable tables and structured `JSON` output for scripting.
-   **Fuzzy Search**: Automatically finds the closest match for common misspellings (e.g., "devlop" -> "develop").
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

To look up a word, simply pass it as an argument to the `ne` command.

**Syntax:**
```bash
./ne [options] <term>
```

**Options:**
-   `--json`, `-j`: Output the result in JSON format.
-   `--full`, `-f`: Show all available data fields for a term.
-   `--dbpath <path>`: Specify a custom path to the `ecdict.bbolt` database file.
-   `--verbose`, `-v`: Enable detailed logging.

## Examples

### Standard Lookup

A standard lookup displays the most common fields in a clean table.

```bash
$ ./ne hello

┌───────────────┬────────────────────────────────────────────────────────────┐
│ term          │ hello                                                      │
├────��──────────┼────────────────────────────────────────────────-───────────┤
│ translation   │ interj. 喂, 嘿                                             │
├───────────────┼────────────────────────────────────────────────────────────┤
│ definition    │ n. an expression of greeting                               │
├───────────────┼────────────────────────────────────────────────────────────┤
│ exchange      │ s:hellos                                                   │
└───────────────┴────────────────────────────────────────────────────────────┘
```

### Fuzzy Search for Misspellings

If you misspell a word, `ne` will automatically search for the closest match.

```bash
$ ./ne devlop

Term 'devlop' not found. Searching for similar terms...
Did you mean 'develop'?

┌───────────────┬────────────────────────────────────────────────────────────┐
│ term          │ develop                                                    │
├───────────────┼────────────────────────────────────────────────────────────┤
│ translation   │ vt. 开发, 发展, 研制, 使成长, 显现出, 冲洗(胶片)             │
│               │ vi. 发育, 生长, 发展, 显露                                 │
├───────────────┼────────────────────────────────────────────────���───────────┤
│ definition    │ grow, progress, unfold, or evolve through a process of     │
│               │ natural growth, differentiation, or a conducive            │
│               │ environment                                                │
├───────────────┼────────────────────────────────────────────────────────────┤
│ exchange      │ d:developed/p:developed/s:develops/i:developing            │
└───────────────┴────────────────────────────────────────────────────────────┘
```

### JSON Output

For scripting or integration with other tools, you can output the full entry as a JSON object.

```bash
$ ./ne hello --json

{
  "term": "hello",
  "data": {
    "audio": "",
    "bnc": "2319",
    "collins": "3",
    "definition": "n. an expression of greeting",
    "detail": "",
    "exchange": "s:hellos",
    "frq": "2238",
    "oxford": "1",
    "phonetic": "hә'lәu",
    "pos": "",
    "tag": "zk gk",
    "translation": "interj. 喂, 嘿"
  }
}
```

## License

This project is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for details.
