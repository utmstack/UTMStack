import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {ActiveDirectoryType} from '../../../active-directory/shared/types/active-directory.type';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../util/request-util';

@Injectable({
  providedIn: 'root'
})
export class ActiveDirectoryDashboardService {

  public resourceUrl = SERVER_API_URL + 'api/ad/dashboard/';

  constructor(private http: HttpClient) {
  }

  getAmountOfAdminsVsUsers(req?: any, top?: any): Observable<HttpResponse<any>> {
    const options = createRequestOption(req);
    return this.http.get<any>(this.resourceUrl + 'amount-of-admins-vs-users', {
      params: options,
      observe: 'response'
    });
  }

  getAmountOfUsersDisabled(req?: any): Observable<HttpResponse<any>> {
    const options = createRequestOption(req);
    return this.http.get<any>(this.resourceUrl + 'amount-of-users-disabled', {params: options, observe: 'response'});
  }

  getAmountOfUsersLockout(req?: any, top?: any): Observable<HttpResponse<any>> {
    const options = createRequestOption(req);
    return this.http.get<any>(this.resourceUrl + 'amount-of-users-lockout', {
      params: options,
      observe: 'response'
    });
  }

  getAmountOfInactiveUsers(req?: any): Observable<HttpResponse<any>> {
    const options = createRequestOption(req);
    return this.http.get<any>(this.resourceUrl + 'amount-of-inactive-users', {
      params: options,
      observe: 'response'
    });
  }

  getAmountOfUsersScaledPrivileges(req?: any): Observable<HttpResponse<any>> {
    const options = createRequestOption(req);
    return this.http.get<any>(this.resourceUrl + 'amount-of-users-scaled-privileges', {
      params: options,
      observe: 'response'
    });
  }

  getAmountOfObjectsInTime(req?: any): Observable<HttpResponse<any>> {
    const options = createRequestOption(req);
    return this.http.get<any>(this.resourceUrl + 'amount-of-events-by-objects-in-time', {
      params: options,
      observe: 'response'
    });
  }

  getAmountOfTrackedUserWithChanges(req?: any): Observable<HttpResponse<any>> {
    const options = createRequestOption(req);
    return this.http.get<any>(this.resourceUrl + 'amount-of-tracked-user-with-changes', {
      params: options,
      observe: 'response'
    });
  }

  getTopMostActiveUsers(req?: any): Observable<HttpResponse<any>> {
    const options = createRequestOption(req);
    return this.http.get<any>(this.resourceUrl + 'top-most-active-users', {
      params: options,
      observe: 'response'
    });
  }

  getTopUserMostChanged(req?: any): Observable<HttpResponse<any>> {
    const options = createRequestOption(req);
    return this.http.get<any>(this.resourceUrl + 'top-users-with-more-changes-made-or-received', {
      params: options,
      observe: 'response'
    });
  }

  getInactiveAdmins(req?: any): Observable<HttpResponse<ActiveDirectoryType[]>> {
    const options = createRequestOption(req);
    return this.http.get<ActiveDirectoryType[]>(this.resourceUrl + 'inactive-admins', {
      params: options,
      observe: 'response'
    });
  }
}
