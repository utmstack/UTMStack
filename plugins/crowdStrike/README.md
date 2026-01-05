# UTMStack Plugin for CrowdStrike Falcon

## Description

UTMStack Plugin for CrowdStrike Falcon is a connector developed in Golang that receives real-time events from `CrowdStrike Falcon Event Streams` and sends them to the `UTMStack` processing server for further processing.

This connector uses a `GRPC` client to communicate with the UTMStack processing server. The client connects through a `Unix socket` that is created in the UTMStack working directory. 

To obtain the events, `CrowdStrike GoFalcon SDK` is used to communicate with the Falcon Event Streams API, providing real-time security event data from your CrowdStrike environment.

Please note that the connector requires valid CrowdStrike Falcon API credentials to run. The connector will not work without proper authentication.

## Configuration

The plugin requires the following configuration parameters:

- **client_id**: OAuth2 Client ID for CrowdStrike Falcon API
- **client_secret**: OAuth2 Client Secret for CrowdStrike Falcon API  
- **member_cid**: (Optional) Member CID for MSSP environments
- **cloud**: Falcon cloud region (us-1, us-2, eu-1, us-gov-1)

## Features

- Real-time event streaming from CrowdStrike Falcon
- Automatic stream discovery and processing
- Error handling and retry mechanisms
- Event batching to optimize performance
- Timeout controls to prevent blocking
- Structured JSON event formatting

## Authentication

This plugin uses OAuth2 authentication with CrowdStrike Falcon API. You need to:

1. Create API client credentials in your CrowdStrike console
2. Ensure the client has the required scopes for Event Streams API
3. Configure the credentials in UTMStack module configuration

For more information on creating API credentials, visit: https://falcon.crowdstrike.com/support/api-clients-and-keys
