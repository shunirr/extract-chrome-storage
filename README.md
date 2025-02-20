# extract-chrome-storage

This is a CLI tool to extract Chrome storage, such as cookies and local storage.

## Requirement

- macOS
- go 1.23.4

## Install

```console
go install github.com/shunirr/extract-chrome-storage@latest
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

### Slack app data

```console
extract-chrome-storage cookie --browser slack --app-store --domain "app.slack.com" | jq .
```

```console
extract-chrome-storage local-storage --browser slack --app-store --domain "app.slack.com" --key "localConfig_v2" | jq .
```
