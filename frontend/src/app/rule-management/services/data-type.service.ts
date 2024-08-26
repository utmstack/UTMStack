import { HttpClient, HttpResponse } from '@angular/common/http';
import {Injectable} from '@angular/core';
import {BehaviorSubject, Observable, of} from 'rxjs';
import {SERVER_API_URL} from '../../app.constants';
import {createRequestOption} from '../../shared/util/request-util';
import {DataType} from '../models/rule.model';
import {switchMap, tap} from "rxjs/operators";
import {InputSourceDataType} from "../../assets-discover/source-data-type-config/input-source-data.type";

const resourceUrl = `${SERVER_API_URL}api/data-types`;

@Injectable({providedIn: 'root'})
export class DataTypeService {

    private typesSubject = new BehaviorSubject<DataType[]>([]);
    type$ = this.typesSubject.asObservable();

    request: {
        page: 0,
        size: 100
    };

    constructor(private http: HttpClient) {
    }

    resetTypes() {
        this.typesSubject.next([]);
    }

    save(body: Partial<DataType>): Observable<HttpResponse<any>> {
        return this.http.post<HttpResponse<any>>(resourceUrl, body, {observe: 'response'});
    }

    update(body: Partial<DataType>): Observable<HttpResponse<any>> {
        return this.http.put<HttpResponse<any>>(resourceUrl, body, {observe: 'response'});
    }

    getById(id: number): Observable<HttpResponse<DataType>> {
      return this.http.get<DataType>(`${resourceUrl}/${id}`, {observe: 'response'});
    }

    delete(id: number): Observable<HttpResponse<any>> {
       return this.http.delete<any>(`${resourceUrl}/${id}`, {observe: 'response'});
    }

    public saveDataType( mode: string, type: Partial<DataType>){
        return mode === 'ADD' ? this.save(type) : this.update(type);
    }

    getAll(req: any): Observable<HttpResponse<DataType[]>> {
        const options = createRequestOption(req);
        return this.http.get<DataType[]>(resourceUrl, {params: options, observe: 'response'})
            .pipe(
                tap(response => {
                    this.typesSubject.next([...this.typesSubject.value, ...response.body]);
                })
            );
    }

  updateInclude(changes: DataType[]): Observable<HttpResponse<DataType>> {
    return this.http.put<DataType>(`${resourceUrl}/include-exclude-list`, changes, {observe: 'response'});
  }
}
