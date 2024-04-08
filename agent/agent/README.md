# UTMStack Agent Service
**UTMStack Agent** is a component of **UTMStack** responsible for the collection of logs and security events. This agent can be installed on both Windows and Linux operating systems. It connects to various services and performs tasks such as log processing, Beats logs reading, Redline service checking, among others. 

## Features
The UTMStack agent performs the following tasks:
- Configures the location of the logs.
- Connects to the Agent Manager and the Log Auth Proxy.
- Creates a client for the Agent Service and the Log Service.
- Processes the logs.
- Reads the Beats logs.
- Checks the Redline service.
- Starts the ping and the incident response stream.
