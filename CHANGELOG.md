# UTMStack 10.6.0 Release Notes
## Bug Fixes
- Reorganized GeoIP database loading into more modular functions for improved maintainability and code readability. Simplified caching, removed unused database function, and restructured rule-handling logic. Improved consistency by standardizing variable names and logging practices.
- Removed unused docker volume configuration for GeoIp.
- Fixed Kernel modules weren't loaded because incorrect function call.

## New Features
- Introduced automatic threat intelligence rules to detect blacklisted ips, hostnames and domains.
