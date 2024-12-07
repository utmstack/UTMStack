# UTMStack Plugin for Microsoft Office 365 (O365)


## Description

UTMStack Plugin for Microsoft Office 365 (O365) is a connector developed in Golang that synchronizes and processes logs from `Office 365` and sends them to a `UTMStack` processing server.

This connector uses a `GRPC` client to communicate with the UTMStack processing server through a `Unix socket` created in the UTMStack working directory.

The `Office 365 API` is used to obtain logs, authenticating via OAuth2 and managing subscriptions for various log types such as `Audit.AzureActiveDirectory`, `Audit.Exchange`, `Audit.General`, `DLP.All`, and `Audit.SharePoint`.

Please note that the connector requires a valid O365 account to run. The connector will not work without an account.
