# extract-chrome-storage

This is a CLI tool for extracting Chrome storage, such as cookies and local storage.

## Requirement

- macOS

## Install

```console
brew install shunirr/extract-chrome-storage/extract-chrome-storage
```

## Usage

```console
extract-chrome-storage --help
```

### Cookie

```console
extract-chrome-storage cookie --domain ".example.com"
```

### Local Storage

```console
extract-chrome-storage local-storage --domain "www.example.com" --key "localStorageData"
```

### Slack app data

```console
extract-chrome-storage cookie --browser slack --app-store --domain ".slack.com" | jq .
```

```console
extract-chrome-storage local-storage --browser slack --app-store --domain "app.slack.com" --key "localConfig_v2" | jq .
```
