import {Component, Input, OnInit} from '@angular/core';
import {FormGroup} from '@angular/forms';
import {Observable, of} from 'rxjs';
import {catchError, map} from 'rxjs/operators';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {ALERT_INDEX_PATTERN, LOG_INDEX_PATTERN} from '../../../../shared/constants/main-index-pattern.constant';
import {FieldDataService} from '../../../../shared/services/elasticsearch/field-data.service';
import {ElasticSearchFieldInfoType} from '../../../../shared/types/elasticsearch/elastic-search-field-info.type';
import {Rule} from '../../../models/rule.model';


@Component({
  selector: 'app-fields-selector',
  templateUrl: './fields-selector.component.html',
  styleUrls: ['./fields-selector.component.css']
})
export class FieldsSelectorComponent implements OnInit {
  @Input() formGroup: FormGroup;
  @Input() rule: Rule;
  @Input() controlName: 'groupBy' | 'deduplicateBy';

  fields$: Observable<ElasticSearchFieldInfoType[]>;
  operators =  [
    {label: 'equals', value: 'eq'},
    {label: 'not equals', value: 'neq'}
  ];

  constructor(private toastService: UtmToastService,
              private fieldDataService: FieldDataService) { }

  ngOnInit() {
    this.fields$ = this.fieldDataService.getFields(LOG_INDEX_PATTERN).pipe(
      map((fields) => fields || []),
      catchError((error) => {
        this.toastService.showError('Error', 'Failed to load fields');
        return of([]);
      })
    );
  }

  addTag(name: string) {
    return { name };
  }
}
