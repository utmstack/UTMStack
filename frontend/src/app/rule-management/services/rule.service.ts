import {HttpClient, HttpHeaders, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable, of} from 'rxjs';
import {catchError, delay} from 'rxjs/operators';
import {SERVER_API_URL} from '../../app.constants';
import {createRequestOption} from '../../shared/util/request-util';
import {Mode, Rule} from '../models/rule.model';

const resourceUrl = `${SERVER_API_URL}api/correlation-rule`;

@Injectable()
export class RuleService {
    rules: Rule[] = [];

    constructor(private http: HttpClient) {
    }

    getRules(req: any): Observable<HttpResponse<Rule[]>> {
        const options = createRequestOption(req);
        return this.http.get<Rule[]>(`${resourceUrl}/search-by-filters`, {params: options, observe: 'response'})
            .pipe(
                catchError((error: any): Observable<HttpResponse<Rule[]>> => {
                    console.error('Error loading rules', error);
                    return of(new HttpResponse({
                        headers: new HttpHeaders({'X-Total-Count': '0'}),
                        body: []
                    }));
                })
            );
    }

    getRuleById(id: number): Observable<HttpResponse<Rule>> {
        return this.http.get<Rule>(`${resourceUrl}/${id}`, {observe: 'response'})
            .pipe(
                catchError((error: any): Observable<HttpResponse<Rule>> => {
                    console.error('Error loading rules', error);
                    return of(new HttpResponse({
                        body: undefined
                    }));
                })
            );
    }

    save(rule: Partial<Rule>): Observable<any> {
        return this.http.post(resourceUrl, rule, {observe: 'response'});
    }

    update(rule: Partial<Rule>): Observable<any> {
        return this.http.put(resourceUrl, rule, {observe: 'response'});
    }

    getFieldValue(params: any) {
        return of(new HttpResponse({
            body: this.getUniqueValues(params.prop)
        })).pipe(
            delay(2000),
            catchError((error: any): Observable<HttpResponse<Array<[string, number]>>> => {
                console.error('Error loading rules', error);
                return of(new HttpResponse({
                    body: undefined
                }));
            })
        );
    }

    public saveRule(mode: Mode, rule: Partial<Rule>) {
        if (mode === 'ADD') {
            return this.save(rule);
        } else {
            return this.update(rule);
        }
    }

    private getUniqueValues(attributeName: string): Array<[string, number]> {
        const uniqueValues: Set<string> = new Set();
        const uniqueValuesArray: Array<[string, number]> = [];

        for (const rule of this.rules) {
            const attributeValue = this.getValueFromPath(rule, attributeName);

            if (Array.isArray(attributeValue)) {
                attributeValue.forEach(a => {
                    if (typeof a === 'object') {
                        uniqueValues.add(a.name);
                    } else {
                        uniqueValues.add(a);
                    }

                });
            }

            if (attributeValue !== undefined && !Array.isArray(attributeValue)) {
                uniqueValues.add(attributeValue.toString());
            }
        }

        Array.from(uniqueValues).forEach((value, index) =>
            uniqueValuesArray.push([value, index]));

        return uniqueValuesArray;
    }

    private getValueFromPath(obj: any, path: string): any {
        const parts = path.split('.');
        let value = obj;
        for (const part of parts) {
            if (value[part] === undefined) {
                return undefined;
            }
            value = value[part];
        }
        return value;
    }
}
