import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {BehaviorSubject, Observable} from 'rxjs';
import {map, tap} from 'rxjs/operators';
import {createRequestOption} from '../../shared/util/request-util';
import {RuleService} from './rule.service';
import {DataType} from "../models/rule.model";

@Injectable()
export class FilterService {

  private resetFilterBehaviorSubject = new BehaviorSubject<boolean>(false);
  resetFilter = this.resetFilterBehaviorSubject.asObservable();

  private filterChangeBehaviorSubject = new BehaviorSubject<{ prop: string, values: any }>(null);
  filterChange = this.filterChangeBehaviorSubject.asObservable();

  private fieldsValuesSubject =
    new BehaviorSubject<Map<string, Array<[string, number]>>>(new Map<string, Array<[string, number]>>());
  fieldsValues$ = this.fieldsValuesSubject.asObservable();

  constructor(private http: HttpClient,
              private ruleService: RuleService) {
  }

  getFieldValues(url: string, params: any, loadMore = false): Observable<Array<[string, number]>> {
    const httpParams = createRequestOption(params);
    return this.http.get<Array<[string, number]>>(url, {params: httpParams, observe: 'response'})
      .pipe(
        tap(response => {
          const filterMap = this.fieldsValuesSubject.value;
          if (loadMore) {
            const values = filterMap.get(params.prop) ? filterMap.get(params.prop) : [];
            filterMap.set(params.prop, [
              ...values,
              ...response.body
            ]);
            this.fieldsValuesSubject.next(filterMap);
          } else {
            filterMap.set(params.prop, response.body);
            this.fieldsValuesSubject.next(filterMap);
          }
        }),
        map(response => response.body)
      );
  }

  resetAllFilters() {
    this.resetFilterBehaviorSubject.next(true);
  }

  onFilterChange(request: any) {
    this.filterChangeBehaviorSubject.next(request);
  }

  resetFieldValues() {
    this.fieldsValuesSubject.next(new Map<string, Array<[string, number]>>());
  }
}

