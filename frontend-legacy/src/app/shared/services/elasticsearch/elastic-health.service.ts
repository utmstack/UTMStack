import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {BehaviorSubject, Observable} from 'rxjs';
import {tap} from 'rxjs/operators';
import {SERVER_API_URL} from '../../../app.constants';
import {ElasticHealthType} from '../../types/elasticsearch/elastic-health.type';

@Injectable({
  providedIn: 'root'
})
export class ElasticHealthService {
  public resourceUrl = SERVER_API_URL + 'api/elasticsearch/cluster/status';

  private healthSubject = new BehaviorSubject<ElasticHealthType>(null); // Observable for health data
  health$ = this.healthSubject.asObservable(); // Observable for components to subscribe to

  constructor(private http: HttpClient) {
  }

  queryHealth(): Observable<HttpResponse<ElasticHealthType>> {
    return this.http.get<ElasticHealthType>(this.resourceUrl, {
      observe: 'response'
    }).pipe(
        tap(response => this.healthSubject.next(response.body))
    );
  }
}
