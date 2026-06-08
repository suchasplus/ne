# Asset Files

This directory contains dictionary data files used by `ne`.

---

## 1. ECDICT — English Dictionary (`ecdict.csv.xz`)

Source: [ECDICT](https://github.com/skywind3000/ECDICT)

The `ecdict.csv` file has been compressed using `xz` to reduce the repository size.

### Decompression

#### On macOS or Linux

```bash
# Decompress the file, the original compressed file will be removed
xz -d ecdict.csv.xz

# If you want to keep the original compressed file, use -k option
# xz -dk ecdict.csv.xz
```

#### On Windows

You may need to install a tool that supports `.xz` files, such as [7-Zip](https://www.7-zip.org/). Once installed, you can right-click the file and extract it.

---

## 2. CC-CEDICT — Chinese-English Dictionary (`cedict_1_0_ts_utf-8_mdbg.txt.gz`)

Source: [MDBG CC-CEDICT](https://www.mdbg.net/chinese/dictionary?page=cc-cedict)

This file provides Chinese (Simplified & Traditional) to English translations. It is distributed in `.gz` format.

### Decompression

#### On macOS or Linux

```bash
# Decompress the file, the original compressed file will be removed
gunzip cedict_1_0_ts_utf-8_mdbg.txt.gz

# If you want to keep the original compressed file, use -k option
# gunzip -k cedict_1_0_ts_utf-8_mdbg.txt.gz
```

#### On Windows

Use [7-Zip](https://www.7-zip.org/) or any tool that supports `.gz` files to extract it.

### File Format

Each non-comment line follows the format:

```
Traditional Simplified [pin1 yin1] /definition 1/definition 2/.../
```

Example:
```
中文 中文 [Zhong1 wen2] /Chinese language/
```

- Lines starting with `#` are comments and should be ignored.
- Pinyin uses tone numbers (1–4, 5 for neutral).
