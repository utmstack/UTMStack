# UTMStack Installer

![Go](https://github.com/UTMStack/installer/workflows/Go/badge.svg)
[![Quality Gate Status](https://qube.atlasinside.com/api/project_badges/measure?project=utmstack_installer&metric=alert_status)](https://qube.atlasinside.com/dashboard?id=utmstack_installer)
[![Bugs](https://qube.atlasinside.com/api/project_badges/measure?project=utmstack_installer&metric=bugs)](https://qube.atlasinside.com/dashboard?id=utmstack_installer)
[![Vulnerabilities](https://qube.atlasinside.com/api/project_badges/measure?project=utmstack_installer&metric=vulnerabilities)](https://qube.atlasinside.com/dashboard?id=utmstack_installer)
[![Security Rating](https://qube.atlasinside.com/api/project_badges/measure?project=utmstack_installer&metric=security_rating)](https://qube.atlasinside.com/dashboard?id=utmstack_installer)
[![Reliability Rating](https://qube.atlasinside.com/api/project_badges/measure?project=utmstack_installer&metric=reliability_rating)](https://qube.atlasinside.com/dashboard?id=utmstack_installer)
[![Maintainability Rating](https://qube.atlasinside.com/api/project_badges/measure?project=utmstack_installer&metric=sqale_rating)](https://qube.atlasinside.com/dashboard?id=utmstack_installer)
[![Lines of Code](https://qube.atlasinside.com/api/project_badges/measure?project=utmstack_installer&metric=ncloc)](https://qube.atlasinside.com/dashboard?id=utmstack_installer)
[![Code Smells](https://qube.atlasinside.com/api/project_badges/measure?project=utmstack_installer&metric=code_smells)](https://qube.atlasinside.com/dashboard?id=utmstack_installer)
[![Duplicated Lines (%)](https://qube.atlasinside.com/api/project_badges/measure?project=utmstack_installer&metric=duplicated_lines_density)](https://qube.atlasinside.com/dashboard?id=utmstack_installer)
[![Technical Debt](https://qube.atlasinside.com/api/project_badges/measure?project=utmstack_installer&metric=sqale_index)](https://qube.atlasinside.com/dashboard?id=utmstack_installer)

## Installation steps

### Preparing for installation
- Download the latest version from https://github.com/UTMStack/installer/releases (You can use `# wget [URL]` to download the installer directly to the server)
- Set execution permissions with `# chmod +x installer`

### Using Terminal User Interface
- Execute the installer without parameters: `# ./installer`

### Using parameterized mode
You can replace the markups of the next examples by real values in order to use the parameterized mode to install UTMStack Master or Probe.
- Master:
`# ./installer master --datadir "[/example/dir]" --db-pass "[ExAmPlEpaSsWoRd]" --fqdn "[server.example.domain]" --customer-name "[Your Business]" --customer-email "[your@email.com]"`

- Probe:
`# ./installer probe --datadir "[/example/dir]" --db-pass "[Master's DB password]" --host "[Master's IP or FQDN]"`
