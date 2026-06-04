import {HttpResponse} from '@angular/common/http';
import {Injectable, SimpleChanges} from '@angular/core';
import {BehaviorSubject, of} from 'rxjs';
import {catchError, filter, finalize, map, switchMap, tap} from 'rxjs/operators';
import {NetScanType} from '../../../../assets-discover/shared/types/net-scan.type';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {UtmAgentManagerService} from '../../../../shared/services/agent/utm-agent-manager.service';
import {AgentType} from '../../../../shared/types/agent/agent.type';
import {
  ModalConfirmationComponent
} from "../../../../shared/components/utm/util/modal-confirmation/modal-confirmation.component";
import {ModalService} from "../../../../core/modal/modal.service";

@Injectable({
  providedIn: 'root'
})
export class AgentSidebarService {

  private request = new BehaviorSubject<any>(null);
  private loading = new BehaviorSubject<boolean>(false);
  private selectedAgent = new BehaviorSubject<AgentType>(null);

  request$ = this.request.asObservable();
  loading$ = this.loading.asObservable();
  selectedAgent$ = this.selectedAgent.asObservable();

  agents$ = this.request$
    .pipe(
      filter(request => !!request),
      tap(() => this.loading.next(true)),
      switchMap((request) => this.utmAgentManagerService.getAgents(request)
        .pipe(
          map((response: HttpResponse<AgentType[]>) => response.body),
          catchError(() => {
            this.toastService.showError('Error', 'Failed to load agents');
            return of([] as AgentType[]);
          }),
          finalize(() => this.loading.next(false))
        )
      ),
    );

  constructor(private utmAgentManagerService: UtmAgentManagerService,
              private toastService: UtmToastService,
              private modalService: ModalService) {}

  loadData(request: any) {
    this.request.next(request);
  }

  selectAgent(agent: AgentType) {
    if (!!agent && !!this.selectedAgent.value && this.selectedAgent.value.hostname !== agent.hostname) {
      this.changeAgent(agent);
    } else {
      this.selectedAgent.next(agent);
    }
  }

  reset() {
    this.request.next(null);
  }

  changeAgent(agent: AgentType): void {
    const modal = this.modalService.open(ModalConfirmationComponent, { centered: true });

    modal.componentInstance.header = 'Active Session Detected';
    modal.componentInstance.message = `You currently have an active console session with the agent "<strong>${this.selectedAgent.value.hostname}</strong>".<br><br>
    Switching to a different agent will terminate the current session and initiate a new one.<br>
    Make sure to save any work before continuing.`;
    modal.componentInstance.confirmBtnText = 'Switch Agent';
    modal.componentInstance.confirmBtnIcon = 'icon-terminal';
    modal.componentInstance.confirmBtnType = 'primary';

    modal.result.then(() => {
      this.selectedAgent.next(agent);
    }).catch(() => {
      console.log('Agent switch canceled by user');
    });
  }

}
