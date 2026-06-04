import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {UtmModulesEnum} from '../../../app-module/shared/enum/utm-module.enum';
import {UtmModulesService} from '../../../app-module/shared/services/utm-modules.service';
import {SERVER_API_URL} from '../../../app.constants';

@Injectable({
  providedIn: 'root'
})
export class ActiveAdModuleActiveService {

  public resourceUrl = SERVER_API_URL + 'api/';

  constructor(private http: HttpClient,
              private moduleService: UtmModulesService) {
  }

  /**
   * Find is ad module is active
   */
  isADModuleActive(): Observable<boolean> {
    return new Observable<boolean>(subscriber => {
      this.moduleService.isActive(UtmModulesEnum.AD_AUDIT).subscribe(response => {
        subscriber.next(response.body);
      });
    });
  }

  // GET /api/ad/utm-ad-trackers/tracking

  getAdChanges(): Observable<HttpResponse<any[]>> {
    return this.http.get<any[]>(this.resourceUrl + 'ad/utm-ad-trackers/tracking', {observe: 'response'});
  }

}
