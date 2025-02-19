# UTMStack 10.5.20 Release Notes
## Bug Fixes
- Fixed the IP location component to accurately determine whether an IP address is public or private.
- Fixed communication from/to agents using secure connections.
- Fixed negative operator evaluation matching on wrong input value due to insufficient checking in correlation engine.
- Reorganized GeoIP database and threat intelligence loading into more modular functions for improved maintainability and code readability. Simplified caching, removed unused database function, and restructured rule-handling logic. Addressed minor variable renames and logging adjustments for consistency.
- Removed unused docker volume configuration for GeoIp.
- Fixed Kernel modules wheren't loaded because incorrect function call

## New Features
- Introduced new standards, sections, dashboards, and visualizations to compliance reports.
- Update ip address to agent.
- Alert generation for down data sources.