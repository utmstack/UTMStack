# Major Changes
- Added new logs authentication proxy.
- Added agents capability to collect remote syslog messages and receive events by HTTP POST.
- Created individual Logstash pipelines for each type of logs.

# Minor Changes
- Using new Bitdefender integration image.
- Create logstash pipelines.yml file.
- Fix input pipeline in Google Cloud and Azure integrations.
- Adding logging in Azure and Google pipelines creation.
- Fix scape error in pipelines creation.
- Add flags to detect errors.
- Check if cloud integrations are enabled and configured.
- Change Bitdefender input port.
- Other minor bugfixes.