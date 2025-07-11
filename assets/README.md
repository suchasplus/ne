# Asset Files

The `ecdict.csv` file has been compressed using `xz` to reduce the repository size.

## Decompression

To use the dictionary file, you need to decompress `ecdict.csv.xz` first.

### On macOS or Linux

You can use the `xz` command-line tool.

```bash
# Decompress the file, the original compressed file will be removed
xz -d ecdict.csv.xz

# If you want to keep the original compressed file, use -k option
# xz -dk ecdict.csv.xz
```

### On Windows

You may need to install a tool that supports `.xz` files, such as [7-Zip](https://www.7-zip.org/). Once installed, you can right-click the file and extract it.
