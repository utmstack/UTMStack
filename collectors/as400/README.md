# UTMStack AS400 Collector

Log collection service for IBM AS/400 (iSeries) systems that integrates with the UTMStack platform for security analysis and event correlation.

## General Description

UTMStack AS400 Collector is a service written in Go that acts as a bridge between IBM AS/400 systems and the UTMStack platform. The service is installed on an intermediate server, connects to multiple locally configured AS/400 systems, collects security logs, and transmits them in real-time to the UTMStack server for analysis.

### Key Features

- **Multi-Server Collection**: Support for multiple AS/400 systems simultaneously
- **Local Configuration**: AS/400 servers are configured in a local file that the collector watches and hot-reloads on change — no restart needed
- **Automatic Reconnection**: Robust handling of disconnections with automatic retries
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
2. Download dependencies (collector Java JAR)
3. Install Java Runtime if necessary
4. Register the collector with UTMStack's Agent Manager
5. Create the local configuration template
6. Create and enable the system service

To update the collector, reinstall it — there is no auto-update service.

## Configuration of AS/400 Servers

AS/400 servers to monitor are configured **locally**, by editing the `as400-config.yaml`
file in the collector's installation directory. The collector watches the file and
reloads automatically on change (no restart required). Install creates a template.

```yaml
servers:
  - tenant: default
    hostname: 10.0.0.5
    userId: QSECOFR
    password: changeme
```

### Parameters per Server

- **Tenant**: Identifier name of the group/server
- **Hostname**: IP address or hostname of the AS/400
- **User ID**: Connection user to the AS/400
- **Password**: Password (automatically encrypted before it is handed to the collector engine)

## License

This project is part of UTMStack. Consult the main project license for more information.
