import {HttpResponse} from '@angular/common/http';
import {Component, EventEmitter, Input, OnDestroy, OnInit, Output} from '@angular/core';
import {AbstractControl, FormArray, FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UUID} from 'angular2-uuid';
import {Observable, Subject} from 'rxjs';
import {concatMap, debounceTime, filter, takeUntil, tap} from 'rxjs/operators';
import {AlertService} from '../../../../../incident-response/shared/services/alert.service';
import {UtmToastService} from '../../../../../shared/alert/utm-toast.service';
import {
  OperatorService
} from '../../../../../shared/components/utm/filters/utm-elastic-filter/shared/util/operator.service';
import {
  ALERT_CATEGORY_DESCRIPTION_FIELD,
  ALERT_DESCRIPTION_FIELD,
  ALERT_FIELDS,
  ALERT_FULL_LOG_FIELD,
  ALERT_ID_FIELD,
  ALERT_INCIDENT_DATE_FIELD,
  ALERT_INCIDENT_FLAG_FIELD,
  ALERT_INCIDENT_MODULE_FIELD,
  ALERT_INCIDENT_OBSERVATION_FIELD,
  ALERT_INCIDENT_USER_FIELD, ALERT_NAME_FIELD,
  ALERT_NOTE_FIELD,
  ALERT_OBSERVATION_FIELD,
  ALERT_REFERENCE_FIELD,
  ALERT_SEVERITY_DESCRIPTION_FIELD,
  ALERT_SEVERITY_FIELD,
  ALERT_SOLUTION_FIELD,
  ALERT_STATUS_FIELD,
  ALERT_STATUS_FIELD_AUTO,
  ALERT_STATUS_LABEL_FIELD,
  ALERT_TAGS_FIELD,
  ALERT_TIMESTAMP_FIELD,
  EVENT_IS_ALERT, FALSE_POSITIVE_OBJECT, LOG_RELATED_ID_EVENT_FIELD
} from '../../../../../shared/constants/alert/alert-field.constant';
import {AUTOMATIC_REVIEW, CLOSED} from '../../../../../shared/constants/alert/alert-status.constant';
import {FILTER_OPERATORS} from '../../../../../shared/constants/filter-operators.const';
import {ALERT_INDEX_PATTERN} from '../../../../../shared/constants/main-index-pattern.constant';
import {ElasticOperatorsEnum} from '../../../../../shared/enums/elastic-operators.enum';
import {ElasticDataService} from '../../../../../shared/services/elasticsearch/elastic-data.service';
import {AlertTags} from '../../../../../shared/types/alert/alert-tag.type';
import {UtmAlertType} from '../../../../../shared/types/alert/utm-alert.type';
import {ElasticFilterType} from '../../../../../shared/types/filter/elastic-filter.type';
import {OperatorsType} from '../../../../../shared/types/filter/operators.type';
import {sanitizeFilters} from '../../../../../shared/util/elastic-filter.util';
import {getValueFromPropertyPath} from '../../../../../shared/util/get-value-object-from-property-path.util';
import {InputClassResolve} from '../../../../../shared/util/input-class-resolve';
import {AlertRuleType} from '../../../alert-rules/alert-rule.type';
import {AlertTagsCreateComponent} from '../../../alert-tags/alert-tags-create/alert-tags-create.component';
import {AlertUpdateTagBehavior} from '../../behavior/alert-update-tag.behavior';
import {AlertManagementService} from '../../services/alert-management.service';
import {AlertRulesService} from '../../services/alert-rules.service';
import {AlertTagService} from '../../services/alert-tag.service';
import {setAlertPropertyValue} from '../../util/alert-util-function';

@Component({
  selector: 'app-alert-rule-create',
  templateUrl: './alert-rule-create.component.html',
  styleUrls: ['./alert-rule-create.component.scss']
})
export class AlertRuleCreateComponent implements OnInit, OnDestroy {
  @Input() alert: UtmAlertType;
  @Input() isForComplete = false;
  @Input() action: 'create' | 'update' | 'select' = 'create';
  @Input() rule: AlertRuleType;
  @Output() ruleAdd = new EventEmitter<AlertRuleType>();
  tags: AlertTags[];
  selected: AlertTags[] = [];
  fields = ALERT_FIELDS;
  operators: OperatorsType[] = FILTER_OPERATORS;
  excludeOperators = [
    ElasticOperatorsEnum.IS_NOT_BETWEEN,
    ElasticOperatorsEnum.IS_BETWEEN,
    ElasticOperatorsEnum.IS_NOT_ONE_OF,
    ElasticOperatorsEnum.IS_ONE_OF,
  ];
  excludeFields = [ALERT_STATUS_FIELD,
    ALERT_TIMESTAMP_FIELD,
    ALERT_STATUS_LABEL_FIELD,
    ALERT_OBSERVATION_FIELD,
    ALERT_NOTE_FIELD,
    ALERT_REFERENCE_FIELD,
    LOG_RELATED_ID_EVENT_FIELD,
    EVENT_IS_ALERT,
    ALERT_INCIDENT_USER_FIELD,
    ALERT_INCIDENT_DATE_FIELD,
    ALERT_INCIDENT_MODULE_FIELD,
    ALERT_INCIDENT_OBSERVATION_FIELD,
    ALERT_SOLUTION_FIELD,
    ALERT_TAGS_FIELD,
    ALERT_SEVERITY_FIELD,
    ALERT_STATUS_FIELD_AUTO,
    ALERT_INCIDENT_FLAG_FIELD,
    ALERT_FULL_LOG_FIELD,
    ALERT_SEVERITY_DESCRIPTION_FIELD,
    ALERT_CATEGORY_DESCRIPTION_FIELD,
    ALERT_ID_FIELD,
    ALERT_DESCRIPTION_FIELD];
  filters: ElasticFilterType[] = [];
  formRule: FormGroup;
  exist: boolean;
  typing = false;
  viewFieldDetail = false;
  uuid = UUID.UUID();
  tagging = false;
  ElasticOperatorsEnum = ElasticOperatorsEnum;
  alerts = [];
  alertRequest = {
    page: 0,
    size: 10,
    sort: '@timestamp,desc',
    filters: [
      {field: ALERT_STATUS_FIELD_AUTO, operator: ElasticOperatorsEnum.IS_NOT, value: AUTOMATIC_REVIEW},
      {field: ALERT_TAGS_FIELD, operator: ElasticOperatorsEnum.IS_NOT, value: FALSE_POSITIVE_OBJECT.tagName},
      {field: '@timestamp', operator: 'IS_BETWEEN', value: ['now-30d', 'now']}
    ],
    dataNature: ALERT_INDEX_PATTERN,
  };
  loading = false;
  refreshingAlert = false;
  alerts$: Observable<any[]>;
  destroy$ = new Subject<void>();

  constructor(public activeModal: NgbActiveModal,
              public inputClass: InputClassResolve,
              private alertRulesService: AlertRulesService,
              private fb: FormBuilder,
              private modalService: NgbModal,
              private utmToastService: UtmToastService,
              private alertUpdateTagBehavior: AlertUpdateTagBehavior,
              private alertServiceManagement: AlertManagementService,
              private alertTagService: AlertTagService,
              private operatorService: OperatorService,
              private elasticDataService: ElasticDataService,
              private alertService: AlertService) {

    this.fields = ALERT_FIELDS.filter(value => !this.excludeFields.includes(value.field));
    this.operators = FILTER_OPERATORS.filter(value => !this.excludeOperators.includes(value.operator));
  }

  ngOnInit() {

    this.initForm();
    this.createDefaultFilters();
    this.getTags();
    this.formRule.get('name').valueChanges.pipe(debounceTime(3000)).subscribe(ruleName => {
      this.searchRule(ruleName);
    });

    if (this.rule) {
      this.filters = [... this.rule.conditions];
      this.selected = this.rule.tags.length > 0 ? [...this.rule.tags] : [];
    }

    this.alerts$ = this.alertService.onRefresh$
      .pipe(
        takeUntil(this.destroy$),
        filter(loading => loading),
        concatMap(() => this.alertService.fetchData(this.alertRequest)))
      .pipe(
        tap((res) => this.loading = !this.loading));
  }

  initForm() {
    this.formRule = this.fb.group({
      id: [this.rule ? this.rule.id : null],
      name: [ this.rule ? this.rule.name : '', Validators.required],
      description: [this.rule ? this.rule.description : '', Validators.required],
      conditions: this.fb.array([], Validators.required),
      tags: [this.rule ? this.rule.tags : null, Validators.required],
    });
  }

  createDefaultFilters() {
    if (this.alert) {
      this.initConditionsFromAlert();
    } else {
      this.initConditionsFromAction();
    }

    this.subscribeToFieldChanges();
  }

  private initConditionsFromAlert() {
    for (const field of this.fields) {
      const value = this.getFieldValue(field.field);
      if (value) {
        const condition = this.buildCondition(field.field, value);
        this.filters.push(condition);
        this.ruleConditions.push(this.createConditionGroup(condition));
      }
    }
  }

  private initConditionsFromAction() {
    if (this.action === 'create') {
      const field = this.fields[0];
      const condition = this.buildCondition(field.field, '');
      this.filters.push(condition);
      this.ruleConditions.push(this.createConditionGroup(condition));
    } else if (this.action === 'update') {
      this.filters = [...this.rule.conditions];
      this.rule.conditions.forEach(condition => {
        this.ruleConditions.push(this.createConditionGroup(condition));
      });
    }
  }

  private buildCondition(field: string, value: any) {
    return {
      field,
      operator: ElasticOperatorsEnum.IS,
      value
    };
  }

  getFieldValue(field: string): any {
    return getValueFromPropertyPath(this.alert, field, null);
  }

  getFieldName(field: string): string {
    return this.fields.filter(value => value.field === field)[0].label;
  }

  saveRule() {
    const request$ = this.action === 'update'
      ? this.alertRulesService.update(this.formRule.value)
      : this.alertRulesService.create(this.formRule.value);

    const tags = this.selected.map(t => t.tagName);

    request$.subscribe(() => {
      const action = this.action === 'update' ? 'updated' : 'created';
      this.utmToastService.showSuccessBottom(`Rule ${this.formRule.get('name').value} ${action} successfully`);

      if (this.alert) {
        const alertId = this.alert.id;
        this.alertServiceManagement.updateAlertTags([alertId], tags, true).subscribe(() => {
          this.alertUpdateTagBehavior.$tagRefresh.next(true);
          this.utmToastService.showSuccessBottom('Tags updated successfully');
          this.tagging = false;

          this.alert = setAlertPropertyValue(ALERT_TAGS_FIELD, tags, this.alert);

          if (this.isFalsePositive()) {
            const observation = `Tag rule ${this.formRule.get('name').value} applied`;
            this.alertServiceManagement.updateAlertStatus([alertId], CLOSED, observation).subscribe(() => {
              this.finalizeRule();
            });
          } else {
            this.finalizeRule();
          }
        });
      } else {
        this.finalizeRule();
      }
    });
  }

  createRule() {
    if (!this.isForComplete) {
      this.alertRulesService.create(this.formRule.value).subscribe(response => {
        this.utmToastService.showSuccessBottom('Rule ' + this.formRule.get('name').value + ' created successfully');
        const tags = this.selected.map(value => value.tagName);
        const alertId = this.alert.id;
        this.alertServiceManagement.updateAlertTags([alertId], tags, true)
          .subscribe(tagsResponse => {
            this.alertUpdateTagBehavior.$tagRefresh.next(true);
            this.utmToastService.showSuccessBottom('Tags added successfully');
            this.tagging = false;
            this.alert = setAlertPropertyValue(ALERT_TAGS_FIELD, tags, this.alert);
            if (this.isFalsePositive()) {
              const observation = 'Tag rule ' + this.formRule.get('name').value + ' applied';
              this.alertServiceManagement.updateAlertStatus([alertId],
                CLOSED, observation).subscribe(al => {
                this.ruleAdd.emit(this.formRule.value);
                this.activeModal.close();
              });
            } else {
              this.ruleAdd.emit(this.formRule.value);
              this.activeModal.close();
            }
          });
      });
    } else {
      this.ruleAdd.emit(this.formRule.value);
      this.activeModal.close();
    }

  }

  deleteFilter(elasticFilterType: ElasticFilterType) {
    const index = this.filters.indexOf(elasticFilterType);
    if (index !== -1) {
      this.filters.splice(index, 1);
      this.formRule.get('conditions').setValue(this.filters);
    }
  }

  addNewTag() {
    const modalRef = this.modalService.open(AlertTagsCreateComponent, {centered: true, size: 'sm'});
    modalRef.componentInstance.addTag.subscribe((tag) => {
      this.utmToastService.showSuccessBottom('Tag ' + tag.name + ' created');
      this.getTags();
      this.selected.push(tag);
      this.formRule.get('tags').setValue(this.selected);
    });
  }


  searchRule(rule: string) {
    this.typing = true;
    this.exist = true;
    setTimeout(() => {
      const req = {
        name: rule
      };
      this.alertRulesService.query(req).subscribe(response => {
        this.exist = response.body.length > 0;
        this.typing = false;
      });
    }, 1000);
  }

  isSelected(tag: AlertTags) {
    return this.selected.findIndex(value => value.id === tag.id) !== -1;
  }

  selectValue( tag: AlertTags) {
    const index = this.selected.findIndex(value => value.id === tag.id);
    if (index === -1) {
      this.selected.push(tag);
    } else {
      this.selected = this.selected.filter(value => value.id !== tag.id);
    }
    this.formRule.get('tags').setValue(this.selected);
  }

  getTags() {
    this.alertTagService.query({page: 0, size: 100}).subscribe(reponse => {
      this.tags = reponse.body;
    });
  }

  isFalsePositive() {
    return this.selected.findIndex(value => value.tagName.includes('False positive')) !== -1;
  }

  getOperators(conditionField: string) {
    const field  = this.fields.find(f => f.field === conditionField);
    if (field) {
      return this.operatorService.getOperators({name: field.field, type: field.type}, this.operators);
    }
    return this.operators;
  }

  onSearch(event: { term: string; items: any[] }) {
    this.alertRequest = {
      ...this.alertRequest,
      filters: [
        ...this.alertRequest.filters.filter(f => f.operator !== ElasticOperatorsEnum.IS_IN_FIELD),
        {field: 'name', operator: ElasticOperatorsEnum.IS_IN_FIELD, value: event.term}
      ]
    };

    this.getAlerts();
  }

  onAlertChange(alert: any){
    this.alert = alert;
    this.filters = [];
    this.formRule.get('conditions').reset();
    this.createDefaultFilters();
  }

  loadMoreAlerts() {
    this.alertRequest = {
      ...this.alertRequest,
      size: this.alertRequest.size + 5
    };
    this.loading = true;
    this.alertService.notifyRefresh(true);
  }


  trackByFn(alert: any) {
    return alert.id;
  }

  getAlerts() {
    this.elasticDataService.search(
      this.alertRequest.page,
      this.alertRequest.size,
      100000000,
      this.alertRequest.dataNature,
      sanitizeFilters(this.alertRequest.filters),
      this.alertRequest.sort).subscribe(
      (res: HttpResponse<any>) => {
        this.alerts = res.body;
        this.loading = false;
        this.refreshingAlert = false;
      },
      (res: HttpResponse<any>) => {
        this.utmToastService.showError('Error', 'An error occurred while listing the alerts. Please try again later.');
      }
    );
  }

  private finalizeRule(): void {
    this.ruleAdd.emit(this.formRule.value);
    this.activeModal.close();
  }

  get ruleConditions() {
    return this.formRule.get('conditions') as FormArray;
  }

  removeRuleCondition(index: number) {
    this.ruleConditions.removeAt(index);
  }

  createConditionGroup(condition: any): FormGroup {
    return this.fb.group({
      field: [condition.field, Validators.required],
      operator: [condition.operator, Validators.required],
      value: [condition.value]
    });
  }

  subscribeToFieldChanges() {
    this.ruleConditions.controls.forEach((group: AbstractControl, index: number) => {
      const fieldControl = group.get('field');
      const operatorControl = group.get('operator');

      fieldControl.valueChanges
        .pipe(takeUntil(this.destroy$))
        .subscribe(newField => {
        const validOperators = this.getOperators(newField).map(op => op.operator);
        if (!validOperators.includes(operatorControl.value)) {
          operatorControl.setValue(null);
        }
      });
    });
  }

  addRuleCondition() {
    this.ruleConditions.push(
      this.createConditionGroup({
        field: this.fields[0].field,
        operator: null,
        value: null
      }));
    this.subscribeToFieldChanges();
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
