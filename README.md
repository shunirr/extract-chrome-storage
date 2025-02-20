# extract-chrome-storage

This is a CLI tool that extracts Chrome's cookie data and Local Storage data.

## Requirement

- macOS
- go 1.23.4

## Install

```console
go install github.com/shunirr/extract-chrome-storage
```

## Usage

```console
extract-chrome-storage --help
```

### Cookie

```console
extract-chrome-storage cookie --domain "www.example.com"
```

### Local Storage

```console
extract-chrome-storage local-storage --domain "www.example.com" --key "localStorageData"
```
