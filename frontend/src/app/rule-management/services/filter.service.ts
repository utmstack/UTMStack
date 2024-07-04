import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {BehaviorSubject, Observable} from 'rxjs';
import {map} from 'rxjs/operators';
import {RuleService} from './rule.service';

@Injectable()
export class FilterService {

    resetFilterBehaviorSubject = new BehaviorSubject<boolean>(false);
    resetFilter = this.resetFilterBehaviorSubject.asObservable();

    constructor(private http: HttpClient,
                private ruleService: RuleService) {
    }

    getFieldValues(url: string, params: any): Observable<Array<[string, number]>> {
        /*return this.http.get<HttpResponse<Array<[string, number]>>>(url, { params })
            .pipe(
                map( response => response.body)
            );*/

        return this.ruleService.getFieldValue(params)
            .pipe(
                map( response => response.body)
        );
    }

    resetAllFilters() {
        this.resetFilterBehaviorSubject.next(true);
    }
}

