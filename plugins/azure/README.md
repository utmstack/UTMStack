# UTMStack Plugin for Microsoft Azure


## Description

UTMStack Plugin for Microsoft Azure is a connector developed in Golang that retrieves logs from `Azure Monitor's Log Analytics` workspace and sends them to the UTMStack processing server for further processing.

This connector uses a `GRPC` client to communicate with the UTMStack processing engine. The client connects through a `Unix socket` located in the UTMStack working directory.

To obtain the logs, the `Azure Go SDK` is used to communicate with the Log Analytics service.
- The azidentity package is used to authenticate with Azure services.
- The azquery package is used for querying the Log Analytics workspace.

### Requirements
**Azure Credentials:**

- Tenant ID
- Client ID
- Client Secret
- Workspace ID

Please note that the connector requires a valid Azure subscription to run. The connector will not work without a subscription.