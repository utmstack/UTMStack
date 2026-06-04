import { HttpClient, HttpResponse } from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {RegexPattern} from '../../models/rule.model';
import {createRequestOption} from '../../../shared/util/request-util';

const resourceUrl = `${SERVER_API_URL}api/regex-pattern`;

@Injectable()
export class PatternManagerService {

    constructor(private http: HttpClient) {}

    save(body: Partial<RegexPattern>): Observable<HttpResponse<any>> {
        return this.http.post<HttpResponse<any>>(resourceUrl, body, {observe: 'response'});
    }

    update(body: Partial<RegexPattern>): Observable<HttpResponse<any>> {
        return this.http.put<HttpResponse<any>>(resourceUrl, body, {observe: 'response'});
    }

    getById(id: number): Observable<HttpResponse<RegexPattern>> {
      return this.http.get<RegexPattern>(`${resourceUrl}/${id}`, {observe: 'response'});
    }

    delete(id: number): Observable<HttpResponse<any>> {
       return this.http.delete<any>(`${resourceUrl}/${id}`, {observe: 'response'});
    }

    public saveRegexPattern( mode: string, type: Partial<RegexPattern>) {
        return mode === 'ADD' ? this.save(type) : this.update(type);
    }

    getAll(req: any): Observable<HttpResponse<RegexPattern[]>> {
        const options = createRequestOption(req);
        return this.http.get<RegexPattern[]>(resourceUrl, {params: options, observe: 'response'});
    }
}
