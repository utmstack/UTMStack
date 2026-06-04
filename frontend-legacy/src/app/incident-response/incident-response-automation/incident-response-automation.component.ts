import {HttpResponse} from '@angular/common/http';
import {Component, OnInit} from '@angular/core';
import {FormBuilder} from '@angular/forms';
import {ActivatedRoute} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {SourcesService} from '../../admin/sources/sources.service';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {
  ModalConfirmationComponent
} from '../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {INCIDENT_AUTOMATION_ALERT_FIELDS} from '../../shared/constants/alert/alert-field.constant';
import {FILTER_OPERATORS} from '../../shared/constants/filter-operators.const';
import {ADMIN_ROLE, USER_ROLE} from '../../shared/constants/global.constant';
import {ITEMS_PER_PAGE} from '../../shared/constants/pagination.constants';
import {SortEvent} from '../../shared/directives/sortable/type/sort-event';
import {ElasticOperatorsEnum} from '../../shared/enums/elastic-operators.enum';
import {TimeFilterType} from '../../shared/types/time-filter.type';
import {IrCreateRuleComponent} from '../shared/component/ir-create-rule/ir-create-rule.component';
import {IncidentResponseRuleService} from '../shared/services/incident-response-rule.service';
import {IncidentRuleType} from '../shared/type/incident-rule.type';

@Component({
  selector: 'app-incident-response-automation',
  templateUrl: './incident-response-automation.component.html',
  styleUrls: ['./incident-response-automation.component.scss']
})
export class IncidentResponseAutomationComponent implements OnInit {
  loading = true;
  rules: IncidentRuleType[];
  range: TimeFilterType;
  totalItems: number;
  itemsPerPage = ITEMS_PER_PAGE;
  request = {
    page: 0,
    size: ITEMS_PER_PAGE,
    sort: '',
    'active.equals': null,
    'agentPlatform.equals': null,
    'createdBy.equals': null
  };
  viewConditions: IncidentRuleType;
  platforms: string[];
  users: string[];

  constructor(private sourcesService: SourcesService,
              private fb: FormBuilder,
              private modalService: NgbModal,
              private utmToastService: UtmToastService,
              private activatedRoute: ActivatedRoute,
              private incidentResponseRuleService: IncidentResponseRuleService) {
    this.activatedRoute.queryParams.subscribe(params => {
      if (params.id) {
        this.request['id.equals'] = params.id;
      }
    });
  }

  ngOnInit() {
    this.getSelect();
    this.getRules();
  }

  getSelect() {
    this.incidentResponseRuleService.getSelectOptions().subscribe(response => {
      if (response.body) {
        this.users = response.body.users;
        this.platforms = response.body.agentPlatform;
      }
    });
  }

  deleteRule(rule) {
    if (rule) {
      this.incidentResponseRuleService.delete(rule.id).subscribe(() => {
        this.utmToastService.showSuccessBottom('Rule ' + rule.name + ' deleted successfully');
        this.getRules();
      });
    }
  }

  deactivateRuleAction(rule: IncidentRuleType, active: boolean) {
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
      this.setActive(rule, active);
    });
  }


  setActive(rule: IncidentRuleType, active: boolean) {
    rule.active = active;
    this.incidentResponseRuleService.update(rule).subscribe(response => {
      this.utmToastService.showSuccessBottom('Incident response automation status changed successfully');
    });
  }


  loadPage(page: number) {
    this.request.page = page - 1;
    this.getRules();
  }

  onItemsPerPageChange($event: number) {
    this.request.size = $event;
    this.getRules();
  }

  getRules() {
    this.incidentResponseRuleService.query(this.request).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  private onSuccess(data, headers) {
    this.totalItems = headers.get('X-Total-Count');
    this.rules = data;
    this.loading = false;
  }

  private onError(error) {
    // this.alertService.error(error.error, error.message, null);
  }

  onSort($event: SortEvent) {
    this.request.sort = $event.column + ',' + $event.direction;
    this.getRules();
  }


  searchByRule($event: string | number) {
    this.request['name.contains'] = $event;
    this.getRules();
  }

  createRule() {
    const modal = this.modalService.open(IrCreateRuleComponent, {size: 'lg', centered: true});
    modal.componentInstance.ruleCreated.subscribe(() => this.getRules());
  }

  getFieldName(field: string): string {
    return INCIDENT_AUTOMATION_ALERT_FIELDS.filter(value => value.field === field)[0].label;
  }

  getFilterName(operator: ElasticOperatorsEnum): string {
    const index = FILTER_OPERATORS.findIndex(value => value.operator === operator);
    if (index !== -1) {
      return FILTER_OPERATORS[index].name;
    } else {
      return operator;
    }
  }

  editRule(rule: IncidentRuleType) {
    const modal = this.modalService.open(IrCreateRuleComponent, {size: 'lg', centered: true});
    modal.componentInstance.rule = rule;
    modal.componentInstance.ruleCreated.subscribe(() => this.getRules());
  }

  onUserChange($event: any) {
    this.request.page = 0;
    this.request['createdBy.equals'] = $event;
    this.getRules();
  }

  onPlatformChange($event: any) {
    this.request.page = 0;
    this.request['agentPlatform.equals'] = $event;
    this.getRules();
  }

  filterByStatus(active: boolean | null) {
    this.request.page = 0;
    this.request['active.equals'] = active;
    this.getRules();
  }

  clearFilters() {
    this.request = {
      page: 0,
      size: ITEMS_PER_PAGE,
      sort: '',
      'active.equals': null,
      'agentPlatform.equals': null,
      'createdBy.equals': null
    };
    this.getRules();
  }

}
