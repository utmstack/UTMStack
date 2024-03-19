# UTMStack Agent Service Dependency: Self-Updater
This project is a crucial dependency for the UTMStack Agent Service. It is specifically designed to handle updates for the UTMStack Updater Service. When the UTMStack Updater Service finds an update for itself, this tool allows it to uninstall and update itself seamlessly.

### Features
Self-Updating: The primary feature of this tool is to enable the UTMStack Updater Service to update itself when an update is found.
Logging: The tool logs all its activities, which can be useful for debugging and tracking its operations.
Error Handling: The tool is designed to handle errors gracefully and log them for further analysis.

### How it Works
The tool first gets the current path of the UTMStack Updater Service. It then configures log saving and reads data from versions.json. It selects the environment and checks if the UTMStack Updater Service is active. If it is, the service is stopped. The tool then updates the UTMStack Updater Service and restarts it.
