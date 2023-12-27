import {Injectable} from '@angular/core';
import {BehaviorSubject, Observable} from 'rxjs';
import {ElasticSearchFieldInfoType} from '../../types/elasticsearch/elastic-search-field-info.type';
import {ElasticSearchIndexService} from './elasticsearch-index.service';
import {LocalFieldService} from './local-field.service';

@Injectable({
  providedIn: 'root'
})
export class FieldDataService {
  $fields = new BehaviorSubject<ElasticSearchFieldInfoType[]>([]);

  constructor(private indexPatternFieldService: ElasticSearchIndexService,
              private localFieldService: LocalFieldService) {
  }

  getFields(pattern: string): Observable<ElasticSearchFieldInfoType[]> {
    return new Observable<ElasticSearchFieldInfoType[]>(subscriber => {
      const fields: ElasticSearchFieldInfoType[] | null = this.localFieldService.getPatternStoredFields(pattern);
      if (!fields || fields.length === 0) {
        this.indexPatternFieldService.getElasticIndexField({indexPattern: pattern})
          .subscribe(response => {
            this.localFieldService.setPatternStoredFields(pattern, response.body);
            subscriber.next(response.body);
          });
      } else {
        subscriber.next(fields);
      }
    });
  }
}
