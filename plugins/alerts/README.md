# UTMStack Alerts Plugin

This plugin handles alert correlation and deduplication for UTMStack. It processes incoming alerts, identifies duplicates, and groups related alerts to maintain a clean and manageable alert environment.

## Features

- **Alert Deduplication**: Automatically detects and skips duplicate alerts based on configured fields.
- **Alert Correlation**: Groups related alerts by linking them to a parent alert using specified fields.
- **Status Management**: Updates the status of parent alerts when new related events occur (e.g., re-opening completed alerts).
- **OpenSearch Integration**: Stores and updates alert documents directly in OpenSearch.
- **Resilience**: Implements retry logic for search and indexing operations and recovers from panics during alert processing.

## Configuration

The plugin requires the following configuration (usually provided via `org.opensearch` configuration):

- `opensearch`: The URL of the OpenSearch cluster to connect to.

## How it works

1. **Initialization**: Connects to OpenSearch and registers itself as a correlation plugin.
2. **Correlation Logic**:
    - Checks if the incoming alert is a duplicate based on `DeduplicateBy` fields.
    - If not a duplicate, searches for a previous alert to link to based on `GroupBy` fields.
    - If a parent alert is found, it updates the parent's status to "Open" if it was previously "Completed".
3. **Indexing**: Creates a new alert document in the `alert` index with metadata like severity, status, and related events.

## Installation

This plugin is part of the UTMStack ecosystem and is typically deployed as a containerized service.

## Development

To build the plugin:

```bash
go build -o alerts-plugin main.go
```

## Dependencies

- `github.com/threatwinds/go-sdk`
- `github.com/tidwall/gjson`
- `google.golang.org/protobuf`
