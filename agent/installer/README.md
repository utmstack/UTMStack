# UTMStack Agent Installer
The UTMStack Agent Installer is a Go application that is part of the UTMStack. It is responsible for various tasks including checking versions, managing configurations, handling dependencies, managing services, and providing utility functions.
This installer is designed to install the services that make up the UTMStack agent. These services include UTMStackAgent, UTMStackRedline, UTMStackUpdater, UTMStackWindowsLogsCollector, and UTMStackModulesLogsCollector.

### Logging
The UTMStack Agent Installer uses a custom logger, referred to as the "beauty logger". This logger is used to print messages to the console in a more readable and aesthetically pleasing format. It also writes error messages and fatal errors to a log file located at path/logs/utmstack_agent_installer.log.

### Error Handling
If the UTMStack Agent Installer encounters an error during execution, it will log the error and then terminate the program. For example, if it fails to get the current path, or if the required ports (9000 and 50051) are not open, it will log an error message and then call log.Fatalf to terminate the program.

### Dependencies
The UTMStack Agent Installer has several dependencies that are managed through the depend package. These dependencies are service binaries that the agent needs to function properly.

### Version Checking
The UTMStack Agent Installer uses the checkversion package to check if the current version of the agent is up to date. If it is not, the agent will update itself to the latest version.
