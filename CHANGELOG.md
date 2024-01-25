
# UTMStack 10.2.1 Release
This update addresses several critical bugs and improves the application's stability and security. 
Among the critical updates are fixes to installation commands on Linux, error-handling enhancements, and updates to security configurations. 

## Summary of the bug fixes included in this release:
- Resolved an issue with the incorrect installation command for Linux environments.
- Fixed a bug where the application could not assign the requested address, leading to operational issues.
- Addressed a crash in the detail view alert when a data source is disconnected.
- Fixed the issue where a down data source showed an open detail in the view.
- Fixed the issue in the security configurations where users with the role ROLE_USER get disconnected.
- Fixed an issue where the properties host and IP of an alert created for a down data source override in the sync process.
- Rectified the emission issue with the logout observable.
