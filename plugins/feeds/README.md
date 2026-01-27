# UTMStack Plugin for ThreadWinds Ingestion


## Description

UTMStack Plugin for ThreadWinds Ingestion is a connector developed in Golang that extracts security entities from `UTMStack incidents and alerts` and sends them to the `ThreadWinds` threat intelligence platform.

This plugin processes incidents from UTMStack, extracts entities (IPs, domains, hashes, emails, etc.) from their associated alerts and events, and ingests them into ThreadWinds for global threat intelligence correlation and enrichment.

The connector automatically registers with ThreadWinds services using the admin email from the UTMStack system. It periodically polls for recent incidents, extracts all relevant entities (network indicators, file hashes, user identities, etc.), builds associations between entities, and sends them to ThreadWinds for analysis.

### Requirements
**ThreadWinds Credentials:**

- API Key
- API Secret

Please note that the connector automatically registers with ThreadWinds using the admin email if credentials are not already configured. The connector requires a valid admin email to run.
