import {ChangeDetectionStrategy, Component, Input} from '@angular/core';

@Component({
  selector: 'app-log-colletor',
  template: `
    <div class="flex-container mt-2 mb-3">
      <ng-select [items]="platforms"
                 bindLabel="name"
                 placeholder="Select platform"
                 [(ngModel)]="selectedPlatform"
                 class="flex-item">
      </ng-select>
      <ng-select [items]="protocols"
                 bindLabel="name"
                 placeholder="Select Protocol"
                 [(ngModel)]="selectedProtocol"
                 class="flex-item">
      </ng-select>
      <ng-select [items]="actions"
                 bindLabel="name"
                 placeholder="Select Action"
                 [(ngModel)]="selectedAction"
                 class="flex-item">
      </ng-select>
    </div>
    <ng-container *ngIf="selectedProtocol && selectedPlatform && selectedAction">
        <span class="font-weight-semibold mb-2">{{selectedPlatform.shell}}</span>
        <app-utm-code-view class="" [code]=command></app-utm-code-view>
    </ng-container>
  `,
  styles: [`
    .flex-container {
      display: flex;
    }

    .flex-item {
      flex-grow: 1;
      margin-right: 10px;
    }
  `],
  changeDetection: ChangeDetectionStrategy.OnPush
})

export class LogCollectorComponent {

  protocols = [
    { id: 1, name: 'TCP' },
    { id: 2, name: 'UDP' },
    { id: 3, name: 'TLS' }
  ];

  actions = [
    { id: 1, name: 'ENABLE', action: 'enable-integration'},
    { id: 2, name: 'DISABLE',  action: 'disable-integration'}
  ];

  platforms = [
    { id: 1, name: 'WINDOWS',
      command: 'sudo bash -c "Start-Process "C:\\Program Files\\UTMStack\\UTMStack Agent\\utmstack_agent_service.exe" -ArgumentList \'ACTION\', \'AGENT\', \'PORT\' -NoNewWindow -Wait\n',
      shell: 'Windows Powershell terminal as “ADMINISTRATOR”'
    },
    { id: 2,
      name: 'LINUX' , command: 'sudo bash -c "/opt/utmstack-linux-agent/utmstack_agent_service ACTION AGENT PORT"',
      shell: 'Linux bash terminal'
    }
  ];

  @Input() agent: string;

  _selectedProtocol: any;
  _selectedPlatform: any;
  _selectedAction: any;
  constructor() {
  }

  get command() {
    this.selectedPlatform.command.replace(/PORT/gi, this.selectedProtocol.name.toLowerCase());
    return this.replaceAll(this.selectedPlatform.command, {
      PORT: this.selectedProtocol.name.toLowerCase(),
      AGENT: this.agent.toLowerCase(),
      ACTION: this.selectedAction.action
    });
  }

  get selectedPlatform() {
    return this._selectedPlatform;
  }

  set selectedPlatform(platform) {
    this._selectedPlatform = platform;
  }

  get selectedProtocol() {
    return this._selectedProtocol;
  }

  set selectedProtocol(protocol) {
     this._selectedProtocol = protocol;
  }

  get selectedAction() {
    return this._selectedAction;
  }

  set selectedAction(action) {
    this._selectedAction = action;
  }

   replaceAll(command, wordsToReplace) {
    return Object.keys(wordsToReplace).reduce(
      (f, s, i) =>
        `${f}`.replace(new RegExp(s, 'ig'), wordsToReplace[s]),
      command
    );
  }
}
