# UTMStack AS400 Collector

Log collection service for IBM AS/400 (iSeries) systems that integrates with the UTMStack platform for security analysis and event correlation.

## General Description

UTMStack AS400 Collector is a service written in Go that acts as a bridge between IBM AS/400 systems and the UTMStack platform. The service is installed on an intermediate server, connects to multiple remotely configured AS/400 systems, collects security logs, and transmits them in real-time to the UTMStack server for analysis.

### Key Features

- **Multi-Server Collection**: Support for multiple AS/400 systems simultaneously
- **Remote Configuration**: Management of AS/400 servers from the UTMStack panel via gRPC streaming
- **Local Persistence**: Temporary log storage in SQLite to ensure delivery in case of network failures
- **Auto-Updates**: Automatic update service included
- **Automatic Reconnection**: Robust handling of disconnections with automatic retries
- **Configurable Retention**: Control of local database size by retention in megabytes
- **Security**: AES encryption for credentials and TLS communication with the server

## Requirements

- **Operating System**: Linux (recommended)
- **Connectivity**: Network access to:
    - UTMStack server (ports 9000, 9001, 50051)
    - AS/400 systems to monitor
- **Java**: Installed automatically during installation
- **Privileges**: Administrator/root permissions to install the service

### Installation Process

1. Verify connectivity with the UTMStack server
2. Download dependencies (collector Java JAR, updater)
3. Install Java Runtime if necessary
4. Register the collector with UTMStack's Agent Manager
5. Create and enable the system service
6. Install the auto-update service

## Configuration of AS/400 Servers

Configuration of AS/400 servers to monitor is performed **from the UTMStack panel**, not locally. The collector automatically receives configuration.

### Parameters per Server

- **Tenant**: Identifier name of the group/server
- **Hostname**: IP address or hostname of the AS/400
- **User ID**: Connection user to the AS/400
- **Password**: Password (automatically encrypted)

## License

This project is part of UTMStack. Consult the main project license for more information.