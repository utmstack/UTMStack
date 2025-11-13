# UTMStack 11.0.0 Release Notes

This is the release notes for **UTMStack v11**, a major update from v10. This version introduces significant improvements and new features aimed at enhancing performance, scalability, and security.

## ⚠️ BREAKING CHANGE - Migration Required

**IMPORTANT:** UTMStack v11 introduces fundamental architectural changes that make it **incompatible with v10**.

- **Direct upgrades from v10 to v11 are NOT supported**
- A **complete migration** is required to move from v10 to v11
- We are currently developing a **migration tool** to facilitate this process
- **Do not attempt to upgrade** your v10 installation to v11 until the migration tool is available

Please contact our support team for guidance on migration planning and timeline.

## Key Highlights

### Performance and Resource Optimization
- **EventProcessor Integration:** Replaced the resource-intensive Logstash with the new **EventProcessor** from Threatwinds, drastically reducing resource usage for data processing.
- **Plugin Architecture:** Introduced a new **plugin system** for official integrations, improving scalability and maintainability.
- **Scalable Processing:** Previous versions required one container per data input. Now, v11 uses two EventProcessor containers—a manager and a worker—allowing each to run its plugins and process logs in parallel. Additional workers can be added as needed to avoid bottlenecks.

### Security Enhancements
- **TLS Improvements:** Strengthened TLS handling across all components.
- **Mandatory Multi-Factor Authentication (MFA):** Added as a required security measure to protect access.

### SOC-AI Enhancements
- **Custom Models Support:** Users can now utilize their own models in SOC-AI integrations, in addition to officially supported models.

### User Interface and Usability
- **UI Overhaul:** Major improvements to visual interfaces for enhanced user experience.
- **SOAR (formerly Incident Response):** Renamed and upgraded to provide automated alert response workflows.
- **Rule Creation Improvements:** Simplified graphical interface for rule creation while maintaining YAML-based configuration options.
- **Log Filter Format Update:** Simplified from complex Logstash syntax to easy-to-use YAML format.

### Centralization and Deployment
- **Central Server:** All instances can now connect to a central server for improved support, enabling remote log submission.
- **Cross-Platform Installation:** Added support for **Red Hat** installations in addition to Ubuntu.
- **Offline On-Premise Installation:** Supported with guided assistance from our engineers for more complex setups.
- **Automatic Updates:** Updates can now be automatically applied from the central server. Users can schedule updates to run at convenient times, ensuring the system remains current without manual checks.

## Summary
UTMStack v11 represents a major leap forward in performance, scalability, security, and usability. The new architecture, plugin system, and central server support ensure that deployments can grow with your organization's needs while simplifying management and operations.

