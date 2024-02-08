
# UTMStack 10.2.2 Release Notes
This update enhances UTMStack's stability, security, and functionality through critical bug fixes and improvements. We have focused on addressing issues reported by our users and identified through our continuous monitoring, improving the overall user experience and the application's resilience against errors and security threats.

## Summary of the bug fixes included in this release:
- Dashboard and Alerts Accuracy: The Overview dashboard displayed incorrect alert values, ensuring accurate monitoring and alerting capabilities.
- Rule History and Filtering: Resolved a problem with rule history filter conditions, improving the accuracy and usability of incident rule history views.
- Integration and Alert Management: Addressed an issue where integration disconnected alerts were triggered too frequently, reducing unnecessary notifications and improving alert management.
- Incident Rules Enhancement: Added a default agent for incident rules, facilitating smoother operation and implementing incident response strategies.
- Log Explorer Stability: Fixed a crash in the log explorer query functionality, enhancing the stability and reliability of log exploration and analysis.
- Incident Response Automation: Improved incident response automation by allowing it to run in default agents, enhancing the efficiency and effectiveness of automated incident responses.
- UI Improvements: Enhancing user interface interaction and usability for Incident response creation.
- Application Stability: Addressed an Auditor module crash issue, improving the application's stability and reliability.
- Logout:  Adding logout observable, ensuring users a more reliable logout process.

## Security and Stability Enhancements:
- Resolved an issue with the incorrect installation command for Linux environments, streamlining the installation process.
- Enhanced error handling for operational issues, improving application resilience.
- Addressed security configuration issues, ensuring users with the role ROLE_USER maintain stable connections.
- Fixed synchronization issues related to alert properties in down data sources, providing accurate alert management.
