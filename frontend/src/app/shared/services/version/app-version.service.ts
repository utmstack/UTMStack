import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {DATE_SETTING_FORMAT_SHORT, DATE_SETTING_TIMEZONE_SHORT} from '../../constants/date-timezone-date.const';
import {AppVersionInfo, VersionInfo} from '../../types/updates/updates.type';
import {VersionType, VersionInfoService} from './version-info.service';


@Injectable({
  providedIn: 'root'
})
export class AppVersionService {
  public resourceUrl = SERVER_API_URL + 'api/';
  public resourceInfo = SERVER_API_URL + 'management/';

  constructor(private http: HttpClient,
              private versionTypeService: VersionInfoService) {
  }

  getVersion(): Observable<HttpResponse<AppVersionInfo>> {
    return this.http.get<AppVersionInfo>(this.resourceUrl + 'info/version', {observe: 'response'});
  }

  getMode(): Observable<HttpResponse<boolean>> {
    return this.http.get<boolean>(this.resourceUrl + 'isInDevelop', {observe: 'response'});
  }

  loadVersionInfo(): Promise<void> {
    return new Promise((resolve, reject) => {
      this.getVersion()
        .subscribe(
          response => {
            const versionInfo = response.body || {} as AppVersionInfo;

            this.versionTypeService.changeAppVersionInfo(versionInfo);
            const version = versionInfo && versionInfo.version || '';
            const versionType = versionInfo.edition.includes('community') || version === ''
              ? VersionType.COMMUNITY
              : VersionType.ENTERPRISE;

            if (versionType !== this.versionTypeService.versionType()) {
              this.versionTypeService.changeVersionType(versionType);
            }
            resolve();
          },
          error => {
            console.error('Unable to load version info', error);
            reject(error);
          }
        );
    });
  }


}
