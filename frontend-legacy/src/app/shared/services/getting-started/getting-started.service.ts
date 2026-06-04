import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {GettingStartedType} from '../../types/getting-started/getting-started.type';
import {createRequestOption} from '../../util/request-util';

@Injectable({
    providedIn: 'root'
})
export class GettingStartedService {

    public resourceUrl = SERVER_API_URL + 'api/utm-getting-started';

    constructor(private http: HttpClient) {
    }

    getSteps(req?: any): Observable<HttpResponse<GettingStartedType[]>> {
        const options = createRequestOption(req);
        return this.http.get<GettingStartedType[]>(this.resourceUrl, {params: options, observe: 'response'});
    }

    initialize(inSaas: boolean): Observable<string> {
        return this.http.post(this.resourceUrl + '/init', {inSaas}, {responseType: 'text'});
    }

    completeStep(step: string): Observable<string> {
        return this.http.get(this.resourceUrl + '/complete?step=' + step, {responseType: 'text'});
    }
}
