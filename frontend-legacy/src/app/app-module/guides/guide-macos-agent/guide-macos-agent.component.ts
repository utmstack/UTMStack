import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';
import {
  FederationConnectionService
} from "../../../app-management/connection-key/shared/services/federation-connection.service";

@Component({
  selector: 'app-guide-macos-agent',
  templateUrl: './guide-macos-agent.component.html',
  styleUrls: ['./guide-macos-agent.component.css']
})
export class GuideMacosAgentComponent implements OnInit {
  @Input() integrationId: number;
  module = UtmModulesEnum;
  @Input() serverId: number;

  platforms = [];
  token: string;

  constructor(private federationConnectionService: FederationConnectionService) { }

  ngOnInit() {
    this.getToken();
  }

  getToken() {
    this.federationConnectionService.getToken().subscribe(response => {
      if (response.body !== null && response.body !== '') {
        this.token = response.body;
      } else {
        this.token = '';
      }
      this.loadPlatforms();
    });
  }

  getInstallCommand(installerName: string): string {
    const ip = window.location.host.includes(':') ? window.location.host.split(':')[0] : window.location.host;

    return `sudo bash -c "mkdir -p /opt/utmstack-macos-agent && \\
curl -k -o /opt/utmstack-macos-agent/${installerName} \\
https://${ip}:9001/private/dependencies/agent/${installerName} && \\
chmod +x /opt/utmstack-macos-agent/${installerName} && \\
/opt/utmstack-macos-agent/${installerName} install ${ip} <secret>${this.token}</secret> yes"`;
  }

  getUninstallCommand(installerName: string): string {
    return `sudo bash -c "/opt/utmstack-macos-agent/${installerName} uninstall || true; \\
launchctl bootout system /Library/LaunchDaemons/UTMStackAgent.plist 2>/dev/null || true; \\
launchctl bootout system /Library/LaunchDaemons/UTMStackUpdater.plist 2>/dev/null || true; \\
rm -f /Library/LaunchDaemons/UTMStackAgent.plist 2>/dev/null || true; \\
rm -f /Library/LaunchDaemons/UTMStackUpdater.plist 2>/dev/null || true; \\
echo 'Removing UTMStack Agent dependencies...' && sleep 10 && \\
rm -rf /opt/utmstack-macos-agent 2>/dev/null || true; \\
echo 'UTMStack Agent removed successfully.'"`;
  }

  private loadPlatforms() {
    const arm64 = 'utmstack_agent_service_darwin_arm64';

    this.platforms = [
      {
        id: 1, name: 'ARM64 (Apple Silicon)',
        install: this.getInstallCommand(arm64),
        uninstall: this.getUninstallCommand(arm64),
        shell: ''
      }
    ];
  }
}
