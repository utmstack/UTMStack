import {HttpResponse} from '@angular/common/http';
import {Component, OnDestroy, OnInit} from '@angular/core';
import {AccountService} from '../../core/auth/account.service';
import {ITEMS_PER_PAGE} from '../../shared/constants/pagination.constants';
import {SortEvent} from '../../shared/directives/sortable/type/sort-event';
import {IncidentOriginTypeEnum} from '../../shared/enums/incident-response/incident-origin-type.enum';
import {IncidentResponseStatusEnum} from '../../shared/enums/incident-response/incident-response-status.enum';
import {UtmDateFormatEnum} from '../../shared/enums/utm-date-format.enum';
import {UtmAgentManagerService} from '../../shared/services/agent/utm-agent-manager.service';
import {AgentCommandType, AgentType} from '../../shared/types/agent/agent.type';
import {IncidentCommandType} from '../../shared/types/incident/incident-command.type';
import {IncidentResponseJobService} from '../shared/services/incident-response-job.service';

@Component({
  selector: 'app-incident-response-view',
  templateUrl: './incident-response-view.component.html',
  styleUrls: ['./incident-response-view.component.scss']
})
export class IncidentResponseViewComponent implements OnInit, OnDestroy {
  manageCommands: any;
  searching = false;
  loading = true;
  page = 1;
  itemsPerPage = ITEMS_PER_PAGE;
  totalItems: number;
  commands: AgentCommandType[] = [];
  requestParams: { pageNumber: number, pageSize: number, searchQuery: string, sortBy: string };
  sortBy = 'createdDate,desc';
  utmDateFormat = UtmDateFormatEnum.UTM_SHORT;
  agentExecution: AgentCommandType;
  statusEnum = IncidentResponseStatusEnum;
  timer: any;
  executeCommand: number;
  incidentOriginTypeEnum = IncidentOriginTypeEnum;
  reasonRun: IncidentCommandType;
  paramMap: Map<string, string> = new Map<string, string>();
  appliedTypes = [
    {label: 'Alert', key: IncidentOriginTypeEnum.ALERT},
    {label: 'Incident', key: IncidentOriginTypeEnum.INCIDENT},
    {label: 'Incident response', key: IncidentOriginTypeEnum.INCIDENT_RESPONSE},
    {label: 'User execution', key: IncidentOriginTypeEnum.USER_EXECUTION},
  ];

  constructor(private incidentResponseJobService: IncidentResponseJobService,
              private agentManagerService: UtmAgentManagerService,
              private accountService: AccountService) {
  }

  ngOnInit() {
    this.requestParams = {
      pageNumber: this.page,
      pageSize: this.itemsPerPage,
      sortBy: 'id,desc',
      searchQuery: ''
    };
    this.getAgentCommandList();
    this.timer = setInterval(() => {
      this.getAgentCommandList();
    }, 30000);

    this.accountService.identity().then(account => {
      this.reasonRun = {
        command: '',
        reason: '',
        originId: account.login,
        originType: IncidentOriginTypeEnum.INCIDENT_RESPONSE
      };
    });
  }

  ngOnDestroy(): void {
    clearInterval(this.timer);
  }

  getAgentCommandList() {
    this.agentManagerService.getAgentCommands(this.requestParams).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  private onSuccess(data, headers) {
    this.totalItems = headers.get('X-Total-Count');
    this.commands = data;
    this.loading = false;
  }

  private onError(error) {
    this.loading = false;
  }

  onSortBy($event: SortEvent) {
    this.requestParams.sortBy = $event.column + ',' + $event.direction;
    this.getAgentCommandList();
  }

  loadPage(page: any) {
    this.requestParams.pageNumber = page;
    this.getAgentCommandList();
  }

  onItemsPerPageChange($event: number) {
    this.itemsPerPage = $event;
    this.requestParams.pageSize = $event;
    this.getAgentCommandList();
  }

  onIrFilterChange($event: any) {
    Object.keys($event).forEach(key => {
      if ($event[key] !== '' && $event[key] !== null) {
        this.requestParams[key] = $event[key];
      } else {
        this.requestParams[key] = undefined;
      }
    });
    this.getAgentCommandList();
  }

  onAgentSelect($event: AgentType) {
    const field = IncidentFilterNamesEnum.AGENT_ID;
    const value = ($event && $event.id) ? $event.id : null;
    this.loadQueryParam(field, value);
  }

  addParamToMap(param: string, value: any) {
    this.paramMap.set(param, value);
  }

  convertParamMapToQueryParam(): string {
    let queryParam = '';
    this.paramMap.forEach((value, key) => {
      if (queryParam === '') {
        queryParam += `${key}=${value}`;
      } else {
        queryParam += `&${key}=${value}`;
      }
    });
    return queryParam;
  }

  deleteParam(key: string) {
    if (this.paramMap.has(key)) {
      this.paramMap.delete(key);
    }
  }

  selectType($event: any) {
    const value = ($event && $event.key) ? $event.key : null;
    this.loadQueryParam(IncidentFilterNamesEnum.ORIGIN_TYPE, value);
  }

  loadQueryParam(field: IncidentFilterNamesEnum, value: any) {
    if (!value) {
      this.deleteParam(field);
    } else {
      this.addParamToMap(field, value);
    }
    this.requestParams.searchQuery = this.convertParamMapToQueryParam();
    this.getAgentCommandList();
  }
}


export enum IncidentFilterNamesEnum {
  AGENT_ID = 'agent_id.Is',
  ORIGIN_TYPE = 'origin_type.Is',
}
