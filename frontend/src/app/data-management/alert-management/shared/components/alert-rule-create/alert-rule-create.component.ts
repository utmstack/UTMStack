import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UUID} from 'angular2-uuid';
import {debounceTime} from 'rxjs/operators';
import {UtmToastService} from '../../../../../shared/alert/utm-toast.service';
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
  ALERT_INCIDENT_USER_FIELD,
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
  EVENT_IS_ALERT,
  LOG_RELATED_ID_EVENT_FIELD
} from '../../../../../shared/constants/alert/alert-field.constant';
import {CLOSED} from '../../../../../shared/constants/alert/alert-status.constant';
import {FILTER_OPERATORS} from '../../../../../shared/constants/filter-operators.const';
import {ElasticOperatorsEnum} from '../../../../../shared/enums/elastic-operators.enum';
import {AlertTags} from '../../../../../shared/types/alert/alert-tag.type';
import {UtmAlertType} from '../../../../../shared/types/alert/utm-alert.type';
import {ElasticFilterType} from '../../../../../shared/types/filter/elastic-filter.type';
import {OperatorsType} from '../../../../../shared/types/filter/operators.type';
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
export class AlertRuleCreateComponent implements OnInit {
  @Input() alert: UtmAlertType;
  @Input() isForComplete = false;
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

  constructor(public activeModal: NgbActiveModal,
              public inputClass: InputClassResolve,
              private alertRulesService: AlertRulesService,
              private fb: FormBuilder,
              private modalService: NgbModal,
              private utmToastService: UtmToastService,
              private alertUpdateTagBehavior: AlertUpdateTagBehavior,
              private alertServiceManagement: AlertManagementService,
              private alertTagService: AlertTagService) {
    this.fields = ALERT_FIELDS.filter(value => !this.excludeFields.includes(value.field));
    this.operators = FILTER_OPERATORS.filter(value => !this.excludeOperators.includes(value.operator));
  }

  ngOnInit() {
    this.getTags();
    this.initForm();
    this.createDefaultFilters();
    this.formRule.get('name').valueChanges.pipe(debounceTime(3000)).subscribe(ruleName => {
      this.searchRule(ruleName);
    });
  }

  initForm() {
    this.formRule = this.fb.group({
      id: [],
      name: ['', Validators.required],
      description: ['', Validators.required],
      conditions: [[], Validators.required],
      tags: [null, Validators.required],
    });
  }

  createDefaultFilters() {
    for (const field of this.fields) {
      if (this.getFieldValue(field.field)) {
        this.filters.push({field: field.field, operator: ElasticOperatorsEnum.IS, value: this.getFieldValue(field.field)});
      }
    }
    this.formRule.get('conditions').setValue(this.filters);
  }

  getFieldValue(field: string): any {
    return getValueFromPropertyPath(this.alert, field, null);
  }

  getFieldName(field: string): string {
    return this.fields.filter(value => value.field === field)[0].label;
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

  deleteFilter(filter: ElasticFilterType) {
    const index = this.filters.indexOf(filter);
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

  selectValue(tag: AlertTags) {
    const index = this.selected.findIndex(value => value.id === tag.id);
    if (index === -1) {
      this.selected.push(tag);
    } else {
      this.selected.splice(index, 1);
    }
    this.formRule.get('tags').setValue(this.selected);
  }

  getTags() {
    this.alertTagService.query({page: 0, size: 100}).subscribe(reponse => {
      this.tags = reponse.body;
      if (this.isForComplete) {
        const index = this.tags.findIndex(value => value.tagName.includes('False positive'));
        if (index !== -1) {
          this.selected.push(this.tags[0]);
          this.formRule.get('tags').setValue(this.selected);
        }
      }
    });
  }


  isFalsePositive() {
    return this.selected.findIndex(value => value.tagName.includes('False positive')) !== -1;
  }
}
