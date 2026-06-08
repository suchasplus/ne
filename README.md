# ne (玩转 ENglish)

A blazingly fast command-line dictionary tool powered by Go and BoltDB.

`ne` (stands for "玩转 ENglish" - "Mastering ENglish") provides instant, offline access to a comprehensive English-Chinese and Chinese-English dictionary directly from your terminal.

## Features

-   **Offline First**: All lookups are performed locally. No internet connection required after initial setup.
-   **Extremely Fast**: Built on Go and using BoltDB, a high-performance key-value store, for near-instantaneous lookups.
-   **Simple & Clean UI**: Results are displayed in a clean, readable table format.
-   **Flexible Output**: Supports both human-readable tables and structured `JSON` output for scripting.
-   **Fuzzy Search**: Automatically finds the closest match for common misspellings (e.g., "devlop" -> "develop").
-   **Comprehensive Data**: Uses the extensive [ECDICT](https://github.com/skywind3000/ECDICT) dictionary data for English, and [CC-CEDICT](https://www.mdbg.net/chinese/dictionary?page=cc-cedict) for Chinese.
-   **Bilingual**: Automatically detects whether your query is Chinese or English — no flags needed.

## Getting Started

### Prerequisites

-   **Go**: Version 1.24 or newer.
-   **xz**: For decompressing `ecdict.csv.xz` (e.g., `xz-utils` on Debian/Ubuntu, `xz` on macOS via Homebrew).
-   **gunzip**: For decompressing `cedict_1_0_ts_utf-8_mdbg.txt.gz` (usually pre-installed on macOS/Linux).

### Installation & Setup

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/suchasplus/ne.git
    cd ne
    ```

2.  **Decompress the Dictionary Data:**
    Both dictionary source files come compressed. Decompress them before building the databases.
    ```bash
    # Decompress the English dictionary
    xz -d assets/ecdict.csv.xz

    # Decompress the Chinese dictionary
    gunzip assets/cedict_1_0_ts_utf-8_mdbg.txt.gz
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

4.  **Build the Databases:**
    Use `kvbuilder` to create the BoltDB databases from the dictionary files.
    ```bash
    # Build the English dictionary database (ecdict.bbolt)
    ./kvbuilder --mode ecdict --csv assets/ecdict.csv

    # Build the Chinese dictionary database (cedict.bbolt)
    ./kvbuilder --mode cedict --csv assets/cedict_1_0_ts_utf-8_mdbg.txt
    ```
    This process may take a minute per database. Both files will be created in `$HOME/.cache/ne/` by default.

## Usage

To look up a word, simply pass it as an argument to the `ne` command.

**Syntax:**
```bash
./ne [options] <term>
```

**Options:**
-   `--json`, `-j`: Output the result in JSON format.
-   `--full`, `-f`: Show all available data fields for a term.
-   `--dbpath <path>`: Specify a custom path to the `ecdict.bbolt` (English) database file.
-   `--cjkdbpath <path>`: Specify a custom path to the `cedict.bbolt` (Chinese) database file.
-   `--verbose`, `-v`: Enable detailed logging.

**Language Detection:**
`ne` automatically detects whether your query contains Chinese characters (CJK Unicode) and routes to the appropriate database. No flags needed.

## Examples

### Chinese Lookup

Query with any Chinese characters (simplified or traditional) to get pinyin and English definitions.

```bash
$ ./ne 中文

┌───────────────┬────────────────────────────────────────────────────────────┐
│ term          │ 中文                                                        │
├───────────────┼────────────────────────────────────────────────────────────┤
│ pinyin        │ Zhong1 wen2                                                │
├───────────────┼────────────────────────────────────────────────────────────┤
│ definitions   │ Chinese language                                           │
└───────────────┴────────────────────────────────────────────────────────────┘
```

Traditional characters also work:
```bash
$ ./ne 漢字
```

### English Standard Lookup

A standard lookup displays the most common fields in a clean table.

```bash
$ ./ne hello

┌───────────────┬────────────────────────────────────────────────────────────┐
│ term          │ hello                                                      │
├───────────────┼────────────────────────────────────────────────-───────────┤
│ translation   │ interj. 喂, 嘿                                              │
├───────────────┼────────────────────────────────────────────────────────────┤
│ definition    │ n. an expression of greeting                               │
├───────────────┼────────────────────────────────────────────────────────────┤
│ exchange      │ s:hellos                                                   │
└───────────────┴────────────────────────────────────────────────────────────┘
```

### Fuzzy Search for Misspellings

If you misspell a word, `ne` will automatically search for similar terms. If multiple suggestions are found, it will list the most likely candidates based on word frequency and length.

```bash
$ ./ne develp

Term 'develp' not found. Searching for similar terms...
Did you mean one of these?
 - devel
 - develop
```

If only one likely candidate is found, it will be displayed directly:
```bash
$ ./ne deveoper

Term 'deveoper' not found. Searching for similar terms...
Did you mean 'developer'?

┌───────────────┬────────────────────────────────────────────────────────────┐
│ term          │ developer                                                  │
# ... (output continues)
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
