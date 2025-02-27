# UTMStack Plugin for Sophos Central


## Description

UTMStack Plugin for Sophos Central is a connector developed in Golang that synchronizes and processes logs from `Sophos Central` and sends them to the UTMStack processing server for further processing.

This connector uses a `GRPC` client to communicate with the UTMStack processing engine. The client connects through a `Unix socket` located in the UTMStack working directory.

To obtain the logs, the Sophos Central API is used to communicate with the SIEM service.

The `ClientID` and `ClientSecret` headers are used to authenticate with Sophos Central services.
The `DataRegion` is used to specify the endpoint for retrieving logs.

### Requirements
**Sophos Central Credentials:**

- ClientID
- ClientSecret

Please note that the connector requires a valid Sophos Central account to run. The connector will not work without an account.
