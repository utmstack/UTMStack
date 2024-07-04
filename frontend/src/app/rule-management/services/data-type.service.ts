import { HttpClient, HttpResponse } from '@angular/common/http';
import {Injectable} from '@angular/core';
import { Observable } from 'rxjs';
import {SERVER_API_URL} from '../../app.constants';
import {createRequestOption} from '../../shared/util/request-util';
import {DataType} from '../models/rule.model';

const resourceUrl = `${SERVER_API_URL}api/data-types`;

@Injectable()
export class DataTypeService {

    request: {
        page: 0,
        size: 100
    };

    constructor(private http: HttpClient) {
    }

    save(body: Partial<DataType>): Observable<HttpResponse<any>> {
      return this.http.post<HttpResponse<any>>(resourceUrl, body, {observe: 'response'});
    }

    getById(id: number): Observable<HttpResponse<DataType>> {
      return this.http.get<DataType>(`${resourceUrl}/${id}`, {observe: 'response'});
    }

    getAll(req: any): Observable<HttpResponse<DataType[]>> {
        const options = createRequestOption(req);
        return this.http.get<DataType[]>(resourceUrl, {params: options, observe: 'response'});
    }
}
