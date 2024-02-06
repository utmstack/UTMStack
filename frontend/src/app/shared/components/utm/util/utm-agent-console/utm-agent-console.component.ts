import {Component, ElementRef, EventEmitter, Input, OnDestroy, OnInit, Output, ViewChild} from '@angular/core';
import {UUID} from 'angular2-uuid';
import {LocalStorageService, SessionStorageService} from 'ngx-webstorage';
import * as SockJS from 'sockjs-client';
import * as Stomp from 'stompjs';
import {SERVER_API_URL, SESSION_AUTH_TOKEN} from '../../../../../app.constants';
import {AccountService} from '../../../../../core/auth/account.service';
import {Account} from '../../../../../core/user/account.model';
import {UtmToastService} from '../../../../alert/utm-toast.service';
import {ITEMS_PER_PAGE} from '../../../../constants/pagination.constants';
import {UtmAgentManagerService} from '../../../../services/agent/utm-agent-manager.service';
import {AgentStatusEnum, AgentType} from '../../../../types/agent/agent.type';
import {IncidentCommandType} from '../../../../types/incident/incident-command.type';
import {replaceBreakLine} from '../../../../util/string-util';
import {INCIDENT_AUTOMATION_ALERT_FIELDS} from "../../../../constants/alert/alert-field.constant";

@Component({
  selector: 'app-utm-agent-console',
  templateUrl: './utm-agent-console.component.html',
  styleUrls: ['./utm-agent-console.component.scss']
})
export class UtmAgentConsoleComponent implements OnInit, OnDestroy {
  @Input() hostname: string;
  @Input() websocketCommand: IncidentCommandType;
  @Output() close = new EventEmitter<boolean>();
  @ViewChild('contentWrapper', {read: ElementRef}) contentWrapper!: ElementRef;
  @ViewChild('commandInput') commandInput!: ElementRef;
  account: Account;
  commandInProgress = false;
  private token: string;
  command = '';
  messages: string[] = [];
  commandHistory: string[] = [];
  private stompClient: any;
  serverUrl = SERVER_API_URL;
  agentStatusEnum = AgentStatusEnum;
  consoleSignal: string;
  agent: AgentType;
  connectionError = false;
  authorize = false;
  pass: string;

  constructor(private localStorage: LocalStorageService,
              private agentManagerService: UtmAgentManagerService,
              private accountService: AccountService,
              private toast: UtmToastService,
              private sessionStorage: SessionStorageService) {
  }

  scrollToBottom() {
    if (this.contentWrapper) {
      const element = this.contentWrapper.nativeElement as HTMLElement;
      element.scrollTop = element.scrollHeight;
    }
  }

  ngOnInit(): void {
    this.consoleSignal = `➜#`;
    this.websocketCommand.command = '';
  }

  startConnection() {
    this.accountService.identity().then(account => {
      this.account = account;
      const sessionToken = this.sessionStorage.retrieve(SESSION_AUTH_TOKEN);
      const localStorageToken = this.localStorage.retrieve(SESSION_AUTH_TOKEN);
      this.token = sessionToken || localStorageToken;
      if (this.token) {
        this.getAgent(this.hostname).then(value => {
          this.agent = value;
          if (this.agent.status === AgentStatusEnum.ONLINE) {
            this.initializeWebSocketConnection();
            this.getCommands(value);
          } else {
            this.connectionError = true;
          }
        });
      }
    });
  }

  getCommands(agent: AgentType) {
    this.agentManagerService.getAgentCommands({
      pageNumber: 1,
      pageSize: ITEMS_PER_PAGE,
      searchQuery: 'agent_id.Is=' + this.agent.id.toString(),
      sortBy: 'id,desc'
    }).subscribe(response => {
      this.commandHistory = response.body.map(value => value.command);
    });
  }

  getAgent(hostname: string): Promise<AgentType> {
    return new Promise<AgentType>((resolve, reject) => {
        this.agentManagerService.getAgent(hostname).subscribe(response => {
          resolve(response.body);
        }, () => reject());
      }
    );
  }

  focusCommandInput() {
    if (this.commandInput) {
      this.commandInput.nativeElement.focus();
    }
  }

  ngOnDestroy(): void {
    if (this.authorize && this.stompClient !== null) {
      this.stompClient.disconnect();
    }
  }

  initializeWebSocketConnection() {
    const ws = new SockJS(this.serverUrl + 'ws?access_token=' + this.token);
    const stompClient = Stomp.over(ws);
    if (stompClient) {
      this.stompClient = stompClient;
      this.stompClient.withCredentials = false;
      this.stompClient.crossDomain = true;
      this.stompClient.connect({}, (frame) => {
        this.openSocket();
        this.focusCommandInput();
      }, this.errorCallBack);
    } else {
      this.connectionError = true;
    }
  }

  errorCallBack(error) {
    console.log('errorCallBack -> ' + error);
    setTimeout(() => {
      this.initializeWebSocketConnection();
    }, 5000);
  }

  sendCommand() {
    this.messages.push(this.consoleSignal + ' ' + this.command);
    this.commandHistory.push(this.command);
    this.commandInProgress = true;
    this.websocketCommand.command = this.command;
    this.stompClient.send('/app/command/' + this.agent.hostname, {}, JSON.stringify(this.websocketCommand));
  }

  openSocket() {
    const subUrl = `/user/topic/${this.agent.hostname}`;
    this.stompClient.subscribe(subUrl, (message) => {
      this.commandInProgress = false;
      this.messages.push(replaceBreakLine(message.body.toString()));
      this.command = '';
      setTimeout(() => {
        this.scrollToBottom();
        this.focusCommandInput();
      }, 150);
    });
  }

  handleCommandHistory(event: KeyboardEvent) {
    const key = event.key;

    if (key === 'ArrowUp') {
      // Handle keyup event
      event.preventDefault();
      this.setCommandFromHistory(true);
    } else if (key === 'ArrowDown') {
      // Handle keydown event
      event.preventDefault();
      this.setCommandFromHistory(false);
    }
  }

  setCommandFromHistory(isPrevious: boolean) {
    if (isPrevious) {
      // Get previous command from history
      const previousCommandIndex = this.commandHistory.length - 1;
      if (previousCommandIndex >= 0) {
        // Set the previous command in your input field or wherever you want to use it
        this.command = this.commandHistory[previousCommandIndex];
        // Remove the previous command from history to avoid duplicates
        this.commandHistory.splice(previousCommandIndex, 1);
      }
    } else {
      // Get next command from history
      const nextCommandIndex = this.commandHistory.length + 1;
      if (nextCommandIndex <= this.commandHistory.length) {
        // Set the next command in your input field or wherever you want to use it
        this.command = this.commandHistory[nextCommandIndex];
        // Remove the next command from history to avoid duplicates
        this.commandHistory.splice(nextCommandIndex, 1);
      }
    }
  }


  closeConsole() {
    this.close.emit(true);
  }

  checkPassword() {
    const uuid = UUID.UUID();
    this.accountService.checkPassword(this.pass, uuid).subscribe(response => {
      if (response.body !== uuid) {
        this.toast.showError('Invalid check UUID', 'UUID to check your password does not match');
      } else {
        this.authorize = true;
        this.startConnection();
      }
    }, () => {
      this.toast.showError('Incorrect Password', 'Please enter your password again');
    });
  }

  insertVariablePlaceholder($event: string) {
    this.command += `$[${$event}]`;
  }
}
