import {Component, Input, OnInit} from '@angular/core';
import {FormGroup} from '@angular/forms';
import {Observable, of} from 'rxjs';
import {catchError, map} from 'rxjs/operators';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {ALERT_INDEX_PATTERN} from '../../../../shared/constants/main-index-pattern.constant';
import {FieldDataService} from '../../../../shared/services/elasticsearch/field-data.service';
import {ElasticSearchFieldInfoType} from '../../../../shared/types/elasticsearch/elastic-search-field-info.type';
import {Rule} from '../../../models/rule.model';


@Component({
  selector: 'app-deduplicate-fields',
  templateUrl: './deduplicate-fields.component.html',
  styleUrls: ['./deduplicate-fields.component.css']
})
export class DeduplicateFieldsComponent implements OnInit {
  @Input() formGroup: FormGroup;
  @Input() rule: Rule;

  fields$: Observable<ElasticSearchFieldInfoType[]>;
  operators =  [
    {label: 'equals', value: 'eq'},
    {label: 'not equals', value: 'neq'}
  ];

  constructor(private toastService: UtmToastService,
              private fieldDataService: FieldDataService) { }

  ngOnInit() {
    this.fields$ = this.fieldDataService.getFields(ALERT_INDEX_PATTERN).pipe(
      map((fields) => fields || []),
      catchError((error) => {
        this.toastService.showError('Error', 'Failed to load fields');
        return of([]);
      })
    );
  }

  addTag(name: string) {
    alert(name)
    return { name };
  }
}
