import {HttpResponse} from '@angular/common/http';
import {Component, Input, OnInit} from '@angular/core';
import {AbstractControl, FormArray, FormBuilder, FormGroup, Validators} from '@angular/forms';
import {of} from 'rxjs';
import {catchError, filter, map, tap} from 'rxjs/operators';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {ALERT_INDEX_PATTERN} from '../../../../shared/constants/main-index-pattern.constant';
import {ElasticOperatorsEnum} from '../../../../shared/enums/elastic-operators.enum';
import {ElasticSearchIndexService} from '../../../../shared/services/elasticsearch/elasticsearch-index.service';
import {ElasticSearchFieldInfoType} from '../../../../shared/types/elasticsearch/elastic-search-field-info.type';
import {IncidentRuleType} from '../../type/incident-rule.type';

@Component({
  selector: 'app-condition-builder',
  templateUrl: './condition-builder.component.html',
  styleUrls: ['./condition-builder.component.scss']
})
export class ConditionBuilderComponent implements OnInit {

  @Input() alert: any;
  @Input() group: FormGroup;
  @Input() rule: IncidentRuleType;
  fields = [];

  constructor(private fb: FormBuilder,
              private elasticSearchIndexService: ElasticSearchIndexService,
              private toastService: UtmToastService) { }

  ngOnInit() {
    this.elasticSearchIndexService.getElasticIndexField( {
      indexPattern: ALERT_INDEX_PATTERN
    }).pipe(
      map((res: HttpResponse<ElasticSearchFieldInfoType[]>) => {
        return res.body.filter(f => f.type !== 'date');
      }),
      tap((body) => this.fields = body ),
      catchError(() => {
        this.toastService.showError('Error', 'An error has occurred while fetching fields');
        return of([]);
      })
    ).subscribe();
  }

  get ruleConditions() {
    return this.group.get('conditions') as FormArray;
  }

  addRuleCondition() {
    const ruleCondition = this.fb.group({
      field: ['', Validators.required],
      value: ['', Validators.required],
      operator: [ElasticOperatorsEnum.IS]
    });

    this.ruleConditions.push(ruleCondition);
  }

  removeRuleCondition(index: number) {
    this.ruleConditions.removeAt(index);
  }

  getConditionFormControl(condition: AbstractControl) {
    return condition as FormGroup;
  }
}
