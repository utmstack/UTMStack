import {AfterViewInit, Component, OnDestroy, OnInit} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {Observable, of} from 'rxjs';
import {catchError, finalize, map, tap} from 'rxjs/operators';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {
  ModalConfirmationComponent
} from '../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {ITEMS_PER_PAGE} from '../../shared/constants/pagination.constants';
import {UtmAlertType} from '../../shared/types/alert/utm-alert.type';
import {TimeFilterType} from '../../shared/types/time-filter.type';
import {IncidentResponseRuleService} from '../shared/services/incident-response-rule.service';
import {IncidentRuleType} from '../shared/type/incident-rule.type';
import {PlaybookService} from '../shared/services/playbook.service';
import {NewPlaybookComponent} from "../shared/component/new-playbook/new-playbook.component";

@Component({
  selector: 'app-playbooks',
  templateUrl: './playbooks.component.html',
  styleUrls: ['./playbooks.component.scss'],
  providers: [PlaybookService]
})
export class PlaybooksComponent implements OnInit, AfterViewInit, OnDestroy {
  loading = true;
  rules: IncidentRuleType[];
  range: TimeFilterType;
  totalItems: number;
  itemsPerPage = ITEMS_PER_PAGE;
  request = {
    page: 0,
    size: 25,
    sort: '',
    'active.equals': null,
    'agentPlatform.equals': null,
    'createdBy.equals': null,
    'systemOwner.equals': false
  };
  platforms: string[];
  users: string[];
  platforms$: Observable<string[]>;
  loadingPlatform = false;

  constructor(private modalService: NgbModal,
              private utmToastService: UtmToastService,
              private incidentResponseRuleService: IncidentResponseRuleService,
              public playbookService: PlaybookService) {

  }

  ngOnInit() {
    this.platforms$ = this.incidentResponseRuleService.getSelectOptions()
      .pipe(
        tap(() => this.loadingPlatform = true),
        map(response => response.body && response.body.users ? response.body.agentPlatform : []),
        catchError(() => {
          this.utmToastService.showError('Error', 'An error occurred while fetching platforms.');
          return of([]);
        }),
        finalize(() => this.loadingPlatform = false));
  }

  ngAfterViewInit(): void {
    this.playbookService.loadData({...this.request});
  }

  deactivateRuleAction(rule: IncidentRuleType) {
    if (rule.active){
      const deleteModalRef = this.modalService.open(ModalConfirmationComponent, {backdrop: 'static', centered: true});
      deleteModalRef.componentInstance.header = 'Deactivate incident response automation';
      deleteModalRef.componentInstance.message = 'Are you sure that you want to deactivate the rule: \n' + rule.name;
      deleteModalRef.componentInstance.confirmBtnText = 'Inactive';
      deleteModalRef.componentInstance.confirmBtnIcon = 'icon-cancel-circle2';
      deleteModalRef.componentInstance.confirmBtnType = 'delete';
      deleteModalRef.componentInstance.textDisplay = 'If you inactive this rule, future alerts' +
        ' will not trigger incident response commands.';
      deleteModalRef.componentInstance.textType = 'warning';
      deleteModalRef.result.then(() => {
        this.setActive(rule, !rule.active);
      });
    } else {
      this.setActive(rule, !rule.active);
    }
  }


  setActive(rule: IncidentRuleType, active: boolean) {
    rule.active = active;
    this.incidentResponseRuleService.update(rule).subscribe(response => {
      this.utmToastService.showSuccessBottom('Incident response automation status changed successfully');
    });
  }


  loadPage(page: number) {
    this.request.page = page - 1;
    this.playbookService.loadData({
      ...this.request,
    });
  }

  onItemsPerPageChange($event: number) {
    this.request.size = $event;
    this.playbookService.loadData({
      ...this.request,
    });
  }

  searchByRule($event: string | number) {
    this.request['name.contains'] = $event;
    this.playbookService.loadData({
      ...this.request,
      page: 0,
    });
  }

  onPlatformChange($event: any) {
    this.playbookService.loadData({
      ...this.request,
      page: 0,
      'agentPlatform.equals': $event
    });
  }


  trackByFn(index: number, playbook: UtmAlertType): any {
    return playbook.id;
  }

  clearFilters() {
    this.request = {
      ...this.request,
      'active.equals': null,
    };
    this.playbookService.loadData(this.request);
  }

  filterByStatus(status: boolean) {
    this.request = {
      ...this.request,
      'active.equals': status,
    };
    this.playbookService.loadData(this.request);
  }

  newPlaybook() {
    this.modalService.open(NewPlaybookComponent, {size: 'lg', backdrop: 'static', centered: true});
  }

  ngOnDestroy(): void {
    this.playbookService.reset();
  }
}
