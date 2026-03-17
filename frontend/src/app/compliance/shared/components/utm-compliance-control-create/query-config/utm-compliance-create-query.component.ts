import { Component, EventEmitter, Input, OnInit, Output } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import {ConsoleOptions} from '../../../../../shared/components/code-editor/code-editor.component';
import {ALERT_INDEX_PATTERN, LOG_INDEX_PATTERN} from '../../../../../shared/constants/main-index-pattern.constant';
import {SqlValidationService} from '../../../../../shared/services/code-editor/sql-validation.service';
import {LocalFieldService} from '../../../../../shared/services/elasticsearch/local-field.service';
import {UtmIndexPattern} from '../../../../../shared/types/index-pattern/utm-index-pattern';
import {
  ComplianceEvaluationRuleEnum,
  ComplianceEvaluationRuleLabels
} from '../../../enums/compliance-evaluation-rule.enum';
import {ComplianceQueryType} from '../../../type/compliance-query.type';

@Component({
  selector: 'app-utm-compliance-create-query',
  templateUrl: './utm-compliance-create-query.component.html',
  styleUrls: ['./utm-compliance-create-query.component.css']
})
export class UtmComplianceCreateQueryComponent implements OnInit {

  @Input() query: ComplianceQueryType = null;
  @Input() indexPatterns: UtmIndexPattern[] = [];
  @Input() indexPatternNames: string[] = [];
  @Output() add = new EventEmitter<ComplianceQueryType>();
  @Output() cancel = new EventEmitter<void>();

  form: FormGroup;
  evaluationRules = Object.values(ComplianceEvaluationRuleEnum).map(rule => ({
    id: rule,
    evaluationRule: ComplianceEvaluationRuleLabels[rule]
  }));

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

    const control = this.form.get('sqlQuery');

    if (control) {
      const cleaned = control.value
        .replace(/(\r\n|\n|\r)/g, ' ')
        .replace(/\s+/g, ' ');
      control.setValue(cleaned);
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

  indexPatternSelected(event: any) {
    const selected = this.indexPatterns.find(p => p.pattern === event);
    if (selected) {
        this.form.get('indexPatternId').setValue(selected.id);
        this.errorMessage = '';
        this.form.get('indexPatternId').markAsPristine();
        this.form.get('indexPatternId').markAsUntouched();

    } else {
      this.errorMessage = 'Invalid index pattern.';
      this.form.get('indexPatternId').setValue('');
      this.form.get('indexPatternId').markAsDirty();
    }
  }
}
