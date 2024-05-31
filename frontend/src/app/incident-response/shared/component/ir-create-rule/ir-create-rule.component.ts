import {Component, ElementRef, EventEmitter, Input, OnInit, Output, ViewChild} from '@angular/core';
import {FormArray, FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {debounceTime} from 'rxjs/operators';
import {UtmNetScanService} from '../../../../assets-discover/shared/services/utm-net-scan.service';
import {NetScanType} from '../../../../assets-discover/shared/types/net-scan.type';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {ALERT_NAME_FIELD, INCIDENT_AUTOMATION_ALERT_FIELDS} from '../../../../shared/constants/alert/alert-field.constant';
import {ALERT_INDEX_PATTERN} from '../../../../shared/constants/main-index-pattern.constant';
import {ElasticOperatorsEnum} from '../../../../shared/enums/elastic-operators.enum';
import {PrefixElementEnum} from '../../../../shared/enums/prefix-element.enum';
import {ElasticSearchIndexService} from '../../../../shared/services/elasticsearch/elasticsearch-index.service';
import {getValueFromPropertyPath} from '../../../../shared/util/get-value-object-from-property-path.util';
import {InputClassResolve} from '../../../../shared/util/input-class-resolve';
import {createElementPrefix, getElementPrefix} from '../../../../shared/util/string-util';
import {IncidentResponseRuleService} from '../../services/incident-response-rule.service';
import {IncidentRuleType} from '../../type/incident-rule.type';

@Component({
  selector: 'app-ir-create-rule',
  templateUrl: './ir-create-rule.component.html',
  styleUrls: ['./ir-create-rule.component.scss']
})
export class IrCreateRuleComponent implements OnInit {
  @Input() alert;
  @Input() rule: IncidentRuleType;
  @ViewChild('autocomplete') autocomplete: ElementRef;
  @Output() ruleCreated = new EventEmitter<boolean>();
  step = 1;
  stepCompleted: number[] = [];
  creating = false;
  formRule: FormGroup;
  agents: NetScanType[];
  alertFields = INCIDENT_AUTOMATION_ALERT_FIELDS;
  platforms: string[];
  command = '';
  exist = true;
  typing = true;
  rulePrefix: string = createElementPrefix(PrefixElementEnum.INCIDENT_RESPONSE_AUTOMATION);
  valuesMap: Map<string, string[]> = new Map();
  maxLength = 512;

  constructor(private incidentResponseRuleService: IncidentResponseRuleService,
              public activeModal: NgbActiveModal,
              private fb: FormBuilder,
              public inputClass: InputClassResolve,
              private utmNetScanService: UtmNetScanService,
              private elasticSearchIndexService: ElasticSearchIndexService,
              private utmToastService: UtmToastService) {

    this.formRule = this.fb.group({
      id: [null],
      name: ['', Validators.required],
      description: ['', Validators.required],
      conditions: this.fb.array([]),
      command: ['', Validators.required],
      active: [true],
      agentType: [false],
      excludedAgents: [[]],
      defaultAgent: [''],
      agentPlatform: ['', Validators.required]
    });
    this.getPlatforms();
  }

  ngOnInit() {
    if (this.rule) {
      this.exist = false;
      this.typing = false;
      this.rulePrefix = getElementPrefix(this.rule.name);
      this.formRule.patchValue(this.rule, {emitEvent: false});
      const name = this.formRule.get('name').value;
      this.formRule.get('name').setValue(this.replacePrefixInName(name));
      for (const condition of this.rule.conditions) {
        this.getValuesForField(condition.field);
        const ruleCondition = this.fb.group({
          field: [condition.field, Validators.required],
          value: [condition.value, Validators.required],
          operator: [condition.operator]
        });
        this.command = this.rule.command;
        this.ruleConditions.push(ruleCondition);
        this.getAgents(this.formRule.get('agentPlatform').value);
        this.formRule.get('excludedAgents').setValue(this.rule.excludedAgents);
        this.formRule.get('agentType').setValue(this.rule.excludedAgents.length === 0 && this.rule.defaultAgent !== '');
        this.formRule.get('defaultAgent').setValue(this.rule.defaultAgent);
      }
    } else if (this.alert) {
      const alertName = this.getValueFromAlert(ALERT_NAME_FIELD);
      const ruleName = this.rulePrefix + alertName;
      this.formRule.get('name').setValue(alertName);
      this.searchRule(ruleName);
      this.addRuleCondition();
      this.ruleConditions.at(0).get('field').setValue(ALERT_NAME_FIELD);
      this.ruleConditions.at(0).get('value').setValue(alertName);
      if (alertName.toLowerCase().includes('window')) {
        this.formRule.get('agentPlatform').setValue('windows');
        this.getAgents('windows');
      }
    } else {
      this.addRuleCondition();
      this.ruleConditions.at(0).get('field').setValue(ALERT_NAME_FIELD);
    }
    this.getValuesForField(ALERT_NAME_FIELD);
    this.formRule.get('name').valueChanges.pipe(debounceTime(1000)).subscribe(value => {
      this.searchRule(this.rulePrefix + value);
    });
  }

  searchRule(rule: string) {
    this.typing = true;
    this.exist = true;
    setTimeout(() => {
      const req = {
        'name.contains': rule
      };
      this.incidentResponseRuleService.query(req).subscribe(response => {
        this.exist = response.body.length > 0;
        this.typing = false;
      });
    }, 1000);
  }

  get ruleConditions() {
    return this.formRule.get('conditions') as FormArray;
  }

  addRuleCondition() {
    const ruleCondition = this.fb.group({
      field: ['', Validators.required],
      value: ['', Validators.required],
      operator: [ElasticOperatorsEnum.IS]
    });

    this.ruleConditions.push(ruleCondition);
  }

  getValueFromAlert(field: string) {
    return getValueFromPropertyPath(this.alert, field, null);
  }

  removeRuleCondition(index: number) {
    this.ruleConditions.removeAt(index);
  }

  backStep() {
    this.step -= 1;
    this.stepCompleted.pop();
  }

  nextStep() {
    if (this.step === 3) {
      this.formRule.get('command').setValue(this.command);
    }
    this.stepCompleted.push(this.step);
    this.step += 1;
  }

  getPlatforms() {
    this.utmNetScanService.getAssetsPlatforms().subscribe(response => {
      this.platforms = response.body;
    });
  }

  getAgents(platform: any) {
    this.formRule.get('excludedAgents').setValue([]);
    this.formRule.get('defaultAgent').setValue('');
    this.utmNetScanService.query({page: 0, size: 10000, agent: true, osPlatform: platform}).subscribe(response => {
      this.agents = response.body;
      if (this.agents.length  === 1) {
        this.formRule.get('excludedAgents').disable();
      }
    });
  }

  onChangeExclude($event: any) {
    const hostnames = $event.map(value => value.assetName);
    this.formRule.get('excludedAgents').setValue(hostnames);
  }

  isCompleted(step: number) {
    return this.stepCompleted.findIndex(value => value === step) !== -1;
  }

  createRule() {
    if (this.rule) {
      this.editRule();
    } else {
      this.saveRule();
    }
  }

  saveRule() {
    const action = 'created';
    const actionError = 'creating';
    this.formRule.get('command').setValue(this.command);
    this.clearAgentTypeSelection();
    this.setNameBeforeSave();
    this.incidentResponseRuleService.create(this.formRule.value).subscribe(() => {
      this.successSaved(action);
    }, () => this.errorSaving(actionError));
  }

  editRule() {
    const action = 'edited';
    const actionError = 'editing';
    this.clearAgentTypeSelection();
    this.formRule.get('command').setValue(this.command);
    this.setNameBeforeSave();
    this.incidentResponseRuleService.update(this.formRule.value).subscribe(() => {
      this.successSaved(action);
    }, () => this.errorSaving(actionError));
  }

  successSaved(action: string) {
    this.utmToastService.showSuccessBottom('Incident response automation ' + action + ' successfully');
    this.ruleCreated.emit(true);
    this.activeModal.close();
  }

  setNameBeforeSave() {
    this.formRule.get('name').setValue(this.rulePrefix + this.formRule.get('name').value);
  }

 clearAgentTypeSelection() {
   if (this.formRule.get('agentType').value) {
     this.formRule.get('excludedAgents').setValue([]);
   } else {
     this.formRule.get('defaultAgent').setValue('');
   }
 }
  errorSaving(action: string) {
    const ruleName: string = this.formRule.get('name').value;
    this.formRule.get('name').setValue(this.replacePrefixInName(ruleName));
    this.utmToastService.showError('Error  ' + action + ' incident automation',
      'An error has occur while trying to ' + action + ' an incident automation, please contact support team');
  }

  insertFieldPlaceholder(field: string) {
    this.command += `$(${field})`;
  }

  onChangeField($event: any, index: number) {
    if (this.alert) {
      const value = this.getValueFromAlert($event.field);
      this.ruleConditions.at(index).get('value').setValue(value);
    } else {
      this.ruleConditions.at(index).get('value').setValue(null);
    }
    if (!this.valuesMap.has($event.field)) {
      // Key doesn't exist, make the request
      this.getValuesForField($event.field);
    }
  }

  isDisable(step: number) {
    switch (step) {
      case 1:
        return !this.formRule.get('name').valid || !this.formRule.get('description').valid || this.exist;
      case 2:
        return !this.formRule.get('agentPlatform').valid || this.ruleConditions.length === 0
            || !this.ruleConditions.valid
            || (this.formRule.get('agentType').value && !this.formRule.get('defaultAgent').value);
      case 3:
        return !this.command || this.command === '';
    }
  }

  replacePrefixInName(name: string) {
    return name.replace(this.rulePrefix, '');
  }

  getFieldValues(key) {
    return this.valuesMap.get(key);
  }

  loadingData(key): boolean {
    return !this.valuesMap.has(key);
  }

  getValuesForField(key) {
    const req = {
      page: 0,
      size: 10,
      indexPattern: ALERT_INDEX_PATTERN,
      keyword: key + '.keyword'
    };
    this.elasticSearchIndexService.getElasticFieldValues(req).subscribe(res => {
      this.valuesMap.set(key, res.body);
    });
  }

  onChangeToggle($event) {
    if ($event) {
      this.formRule.get('excludedAgents').setValue([]);
    } else {
      this.formRule.get('defaultAgent').setValue('');
    }
    this.formRule.get('agentType').setValue($event);
  }

  insertVariablePlaceholder($event: string) {
    this.command += `$[${$event}]`;
  }


}
