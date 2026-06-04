import { HttpClient, HttpResponse } from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../../shared/util/request-util';
import {Asset} from '../../models/rule.model';

const resourceUrl = `${SERVER_API_URL}api/tenant-config`;

@Injectable()
export class AssetManagerService {

    constructor(private http: HttpClient) {}

    save(body: Partial<Asset>): Observable<HttpResponse<any>> {
        return this.http.post<HttpResponse<any>>(resourceUrl, body, {observe: 'response'});
    }

    update(body: Partial<Asset>): Observable<HttpResponse<any>> {
        return this.http.put<HttpResponse<any>>(resourceUrl, body, {observe: 'response'});
    }

    getById(id: number): Observable<HttpResponse<Asset>> {
      return this.http.get<Asset>(`${resourceUrl}/${id}`, {observe: 'response'});
    }

    delete(id: number): Observable<HttpResponse<any>> {
       return this.http.delete<any>(`${resourceUrl}/${id}`, {observe: 'response'});
    }

    public saveAsset( mode: string, type: Partial<Asset>) {
        return mode === 'ADD' ? this.save(type) : this.update(type);
    }

    getAll(req: any): Observable<HttpResponse<Asset[]>> {
        const options = createRequestOption(req);
        return this.http.get<Asset[]>(resourceUrl, {params: options, observe: 'response'});
    }
}
