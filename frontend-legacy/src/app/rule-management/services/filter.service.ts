import {HttpClient, HttpParams} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {BehaviorSubject, Observable} from 'rxjs';
import {RefreshDataService} from '../../shared/services/util/refresh-data.service';
import {createRequestOption} from '../../shared/util/request-util';
import {RuleService} from './rule.service';

@Injectable()
export class FilterService extends RefreshDataService<string, Array<[string, number]>> {

  private resetFilterBehaviorSubject = new BehaviorSubject<boolean>(false);
  resetFilter = this.resetFilterBehaviorSubject.asObservable();

  private filterChangeBehaviorSubject = new BehaviorSubject<{ prop: string, values: any }>(null);
  filterChange = this.filterChangeBehaviorSubject.asObservable();

  private fieldsValuesSubject =
    new BehaviorSubject<Map<string, Array<[string, number]>>>(new Map<string, Array<[string, number]>>());
  fieldsValues$ = this.fieldsValuesSubject.asObservable();

  constructor(private http: HttpClient,
              private ruleService: RuleService) {
    super();
  }

  fetchData(request: any): Observable<Array<[string, number]>> {
    return this.getFieldValues(request.url, createRequestOption(request.params));
  }

  private getFieldValues(url: string, options: HttpParams): Observable<Array<[string, number]>> {
    return this.http.get<Array<[string, number]>>(url, {params: options});
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

