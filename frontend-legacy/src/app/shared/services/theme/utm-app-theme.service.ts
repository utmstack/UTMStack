import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {AppThemeType} from '../../types/theme/app-theme.type';
import {createRequestOption} from '../../util/request-util';

@Injectable({
  providedIn: 'root'
})
export class UtmAppThemeService {
  public resourceUrl = SERVER_API_URL + 'api/images';

  constructor(private http: HttpClient) {
  }

  update(theme: AppThemeType): Observable<HttpResponse<AppThemeType>> {
    return this.http.put<AppThemeType>(this.resourceUrl, theme, {observe: 'response'});
  }

  getTheme(req?: any): Observable<HttpResponse<AppThemeType[]>> {
    const options = createRequestOption(req);
    return this.http.get<AppThemeType[]>(this.resourceUrl + '/all', {params: options, observe: 'response'});
  }

  resetTheme(req?: any): Observable<HttpResponse<AppThemeType[]>> {
    const options = createRequestOption(req);
    return this.http.get<AppThemeType[]>(this.resourceUrl + '/reset', {params: options, observe: 'response'});
  }

  getImageTheme(shortName: string): Observable<HttpResponse<AppThemeType>> {
    return this.http.get<AppThemeType>(`${this.resourceUrl}/${shortName}`, {observe: 'response'});
  }

}
