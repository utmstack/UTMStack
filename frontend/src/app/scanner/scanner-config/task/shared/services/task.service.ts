import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../../app.constants';
import {createRequestOption} from '../../../../../shared/util/request-util';
import {ITask} from '../../../../shared/model/task.model';

@Injectable({
  providedIn: 'root'
})
export class TaskService {
  private resourceUrl = SERVER_API_URL + 'api/openvas/tasks';

  constructor(private http: HttpClient) {
  }

  create(task: any): Observable<HttpResponse<ITask>> {
    return this.http.post<any>(this.resourceUrl, task, {observe: 'response'});
  }

  update(task: ITask): Observable<HttpResponse<ITask>> {
    return this.http.put<ITask>(this.resourceUrl, task, {observe: 'response'});
  }

  find(task: string): Observable<HttpResponse<ITask>> {
    return this.http.get<ITask>(`${this.resourceUrl}/${task}`, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<ITask[]>> {
    const options = createRequestOption(req);
    return this.http.get<ITask[]>(this.resourceUrl + '/get-by-filters', {params: options, observe: 'response'});
  }

  delete(login: string): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${login}`, {observe: 'response'});
  }

  start(taskId): Observable<HttpResponse<any>> {
    return this.http.post<ITask>(this.resourceUrl + '/start', taskId, {observe: 'response'});
  }

  stop(taskId): Observable<HttpResponse<any>> {
    return this.http.post<ITask>(this.resourceUrl + '/stop', taskId, {observe: 'response'});
  }

  resume(taskId): Observable<HttpResponse<any>> {
    return this.http.post<ITask>(this.resourceUrl + '/resume', taskId, {observe: 'response'});
  }

}
