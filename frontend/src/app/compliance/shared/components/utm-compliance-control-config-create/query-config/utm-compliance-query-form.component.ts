import { Component, EventEmitter, Input, OnInit, Output } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import {UtmIndexPattern} from '../../../../../shared/types/index-pattern/utm-index-pattern';
import {UtmComplianceQueryConfigType} from '../../../type/compliance-query-config.type';
import {HttpResponse} from "@angular/common/http";
import {IndexPatternService} from "../../../../../shared/services/elasticsearch/index-pattern.service";

@Component({
  selector: 'app-utm-compliance-query-form',
  templateUrl: './utm-compliance-query-form.component.html'
})
export class UtmComplianceQueryFormComponent implements OnInit {

  @Input() query: UtmComplianceQueryConfigType = null;

  @Output() save = new EventEmitter<UtmComplianceQueryConfigType>();
  @Output() cancel = new EventEmitter<void>();

  form: FormGroup;
  pattern: UtmIndexPattern;
  patterns: UtmIndexPattern[];
  indexPatternList = [];

  constructor(private fb: FormBuilder,
              private indexPatternService: IndexPatternService) {}

  ngOnInit() {
    this.getIndexPatterns('init');
    const q = this.query || {};

    this.form = this.fb.group({
      id: [q.id || null],
      name: [q.name || '', Validators.required],
      queryDescription: [q.queryDescription || '', Validators.required],
      sqlQuery: [q.sqlQuery || '', Validators.required],
      evaluationRule: [q.evaluationRule || null, Validators.required],
      indexPatternId: [q.indexPatternId || null, Validators.required],
      controlConfigId: [q.controlConfigId || null]
    });
  }

  submit() {
    if (this.form.invalid) {
      return;
    }
    this.save.emit(this.form.value);
  }

  cancelEdit() {
    this.cancel.emit();
  }

  //TODO: ELENA revisar de aqui para abajo
  onIndexPatternChange($event: any) {
    this.pattern = $event;
    //this.indexPatternChange.emit($event);
  }

  getIndexPatterns(init?: string) {
    const req = {
      page: 0,
      size: 1000,
      sort: 'id,asc',
      'isActive.equals': true,
    };
    this.indexPatternService.query(req).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers, init),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  private onSuccess(data, headers, init) {
    this.patterns = data;
    this.indexPatternList = this.getListPatterns();
    if (init) {
      this.pattern = this.patterns[0];
    }
  }

  getListPatterns(){
    return this.patterns.map(pattern => ({ id: pattern.id, name: pattern.pattern, selected: this.pattern.id == pattern.id }));
  }

  private onError(error) {
    // this.alertService.error(error.error, error.message, null);
  }
}
