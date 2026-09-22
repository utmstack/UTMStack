# UTMStack Plugin for ThreadWinds Ingestion


## Description

UTMStack Plugin for ThreadWinds Ingestion is a connector developed in Golang that extracts security entities from `UTMStack incidents and alerts` and sends them to the `ThreadWinds` threat intelligence platform.

It periodically polls for recent incidents, extracts the relevant entities (IPs, domains, hashes, emails, etc.) from their associated alerts and events, builds associations between them, and contributes them to ThreadWinds for global threat intelligence correlation and enrichment.

## Configuration

The functionality is enabled or disabled via the on/off switch provided in the UI. When disabled (or if the file does not exist), the plugin performs no action. The installer creates the file with the option enabled by default for new installations and does not modify it subsequently.

## Requirements

The plugin requires the instance configuration provided by the installer. If that configuration is not available, the plugin will not start.