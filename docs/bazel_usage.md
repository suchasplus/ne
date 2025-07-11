# Bazel Usage Guide for the 'ne' Project

This document provides a quick reference for common Bazel commands used in this project.

## Building

### Build All Targets

To build all binaries and libraries in the project:

```bash
bazel build //...
```

This command compiles the `ne` and `kvbuilder` tools, placing the output in the `bazel-bin` directory.

## Running

### Run the `ne` Dictionary Tool

To run the `ne` tool directly using Bazel:

```bash
bazel run //cmd/ne -- <word>
```

For example, to look up the word "hello":

```bash
bazel run //cmd/ne -- hello
```

To see its help message:

```bash
bazel run //cmd/ne -- --help
```

### Run the `kvbuilder` Database Tool

To run the `kvbuilder` tool:

```bash
bazel run //cmd/kvbuilder -- --csv /path/to/your.csv --dbpath /path/to/your.db
```

To see its help message:

```bash
bazel run //cmd/kvbuilder -- --help
```

## Testing

### Run All Tests

To run all tests in the project:

```bash
bazel test //...
```

## Dependency and Build File Management (Gazelle)

### Update BUILD.bazel Files

If you add or remove Go source files, you need to update the `BUILD.bazel` files. Run `gazelle` for this:

```bash
bazel run //:gazelle
```

### Update Go Dependencies

To sync Go module dependencies from `go.mod` into your `WORKSPACE` file:

```bash
bazel run //:gazelle -- update-repos -from_file=go.mod
```
