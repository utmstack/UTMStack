# UTMStack Plugin for Amazon Web Services (AWS)


## Description

UTMStack Plugin for Amazon Web Services (AWS) is a connector developed in Golang that receives logs from `AWS CloudWatch` and sends them to the `UTMStack` processing server for further processing.

This connector uses a `GRPC` client to communicate with the UTMStack processing server. The client connects through a `Unix socket` that is created in the UTMStack working directory. 

To obtain the logs, `AWS Go SDK` is used to communicate with the CloudWatch Logs service.

Please note that the connector requires a valid AWS account to run. The connector will not work without an account.

