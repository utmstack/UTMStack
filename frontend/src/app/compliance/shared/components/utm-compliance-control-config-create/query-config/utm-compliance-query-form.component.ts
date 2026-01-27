import { Component, EventEmitter, Input, OnInit, Output } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import {ConsoleOptions} from '../../../../../shared/components/code-editor/code-editor.component';
import {ALERT_INDEX_PATTERN, LOG_INDEX_PATTERN} from '../../../../../shared/constants/main-index-pattern.constant';
import {SqlValidationService} from '../../../../../shared/services/code-editor/sql-validation.service';
import {LocalFieldService} from '../../../../../shared/services/elasticsearch/local-field.service';
import {UtmIndexPattern} from '../../../../../shared/types/index-pattern/utm-index-pattern';
import {ComplianceEvaluationRuleEnum} from '../../../enums/compliance-evaluation-rule.enum';
import {UtmComplianceQueryConfigType} from '../../../type/compliance-query-config.type';

@Component({
  selector: 'app-utm-compliance-query-form',
  templateUrl: './utm-compliance-query-form.component.html',
  styleUrls: ['./utm-compliance-query-form.component.css']
})
export class UtmComplianceQueryFormComponent implements OnInit {

  @Input() query: UtmComplianceQueryConfigType = null;
  @Input() indexPatterns: UtmIndexPattern[] = [];
  @Input() indexPatternNames: string[] = [];
  @Output() add = new EventEmitter<UtmComplianceQueryConfigType>();
  @Output() cancel = new EventEmitter<void>();

  form: FormGroup;
  evaluationRules = Object.values(ComplianceEvaluationRuleEnum);

  codeEditorOptions: ConsoleOptions = {lineNumbers: 'off'};
  errorMessage = '';

  constructor(private fb: FormBuilder,
              private sqlValidationService: SqlValidationService,
              private localFieldService: LocalFieldService) {}

  ngOnInit() {
    const q = this.query || {};

    this.form = this.fb.group({
      id: [q.id || null],
      queryName: [q.queryName || '', [Validators.required, Validators.minLength(10), Validators.maxLength(200)]],
      queryDescription: [q.queryDescription || '', [Validators.required, Validators.maxLength(2000)]],
      sqlQuery: [q.sqlQuery || '', [Validators.required, Validators.maxLength(2000)]],
      evaluationRule: [q.evaluationRule || null, [Validators.required]],
      ruleValue: [q.ruleValue || '', ],
      indexPatternId: [q.indexPatternId || null, [Validators.required]],
      controlConfigId: [q.controlConfigId || null]
    });

    this.form.get('evaluationRule').valueChanges.subscribe(rule => {
      const ruleValueControl = this.form.get('ruleValue');

      if (rule !== null && rule !== 'NO_HITS_ALLOWED') {
        ruleValueControl.setValidators([
          Validators.required,
          Validators.min(1),
          Validators.pattern(/^[1-9]\d*$/)
        ]);
      } else {
        ruleValueControl.clearValidators();
        ruleValueControl.setValue(null);
      }

      ruleValueControl.updateValueAndValidity();
    });
  }

  submit() {
    if (this.form.invalid) {
      return;
    }

    const sql = this.form.get('sqlQuery').value;
    this.errorMessage = this.sqlValidationService.validateSqlQuery(sql);
    if (this.errorMessage) {
      return;
    }

    this.add.emit(this.form.value);
    this.form.reset();
  }

  cancelEdit() {
    this.cancel.emit();
  }

  loadFieldNames() {
    return [
      ...this.localFieldService.getPatternStoredFields(ALERT_INDEX_PATTERN).map(f => f.name),
      ...this.localFieldService.getPatternStoredFields(LOG_INDEX_PATTERN).map(f => f.name)
    ];
  }
}
