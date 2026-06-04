import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormArray, FormGroup} from '@angular/forms';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {INCIDENT_AUTOMATION_ALERT_FIELDS} from '../../../../shared/constants/alert/alert-field.constant';
import {FILTER_OPERATORS} from '../../../../shared/constants/filter-operators.const';
import {ElasticOperatorsEnum} from '../../../../shared/enums/elastic-operators.enum';
import {IncidentResponseAssetService} from '../../services/incident-response-asset.service';
import {IncidentResponseJobService} from '../../services/incident-response-job.service';

@Component({
  selector: 'app-ir-summary',
  templateUrl: './ir-summary.component.html',
  styleUrls: ['./ir-summary.component.scss']
})
export class IrSummaryComponent implements OnInit {
  @Input() formRule: FormGroup;
  @Output() jobCreated = new EventEmitter<string>();
  alertFields = INCIDENT_AUTOMATION_ALERT_FIELDS;


  constructor(private incidentResponseAssetService: IncidentResponseAssetService,
              public utmToastService: UtmToastService,
              private incidentResponseJobService: IncidentResponseJobService) {
  }

  ngOnInit() {
  }

  get automationName() {
    return this.formRule.get('name').value;
  }

  get platform() {
    return this.formRule.get('agentPlatform').value;
  }

  get defaultAgent() {
    return this.formRule.get('defaultAgent').value;
  }

  get command() {
    return this.formRule.get('command').value;
  }

  get agentType() {
    return this.formRule.get('agentType').value;
  }

   get conditions() {
    return (this.formRule.get('conditions') as FormArray).value;
  }

  get excludedAgents() {
    return this.formRule.get('excludedAgents').value;
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

}
