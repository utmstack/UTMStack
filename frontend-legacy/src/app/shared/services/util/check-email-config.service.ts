import { Injectable } from '@angular/core';
import {Router} from '@angular/router';
import {map} from 'rxjs/operators';
import {ModalService} from '../../../core/modal/modal.service';
import {
  EmailSettingNotificationComponent
} from '../../components/email-setting-notification/email-setting-notification.component';
import {UtmConfigParamsService} from '../config/utm-config-params.service';

export enum ParamShortType {
  Incident= 'utmstack.alert.addressToNotifyIncidents',
  Alert =  'utmstack.alert.addressToNotifyAlerts'
}

@Injectable({
  providedIn: 'root'
})
export class CheckEmailConfigService {

  constructor(private utmConfigParamsService: UtmConfigParamsService,
              private modalService: ModalService,
              private router: Router) { }

  check(paramShort: ParamShortType) {
    this.utmConfigParamsService.query({
      page: 0,
      size: 10000,
      'sectionId.equals': 4,
      'confParamShort.equals': paramShort,
      sort: 'id,asc'
    }).pipe(
        map(res =>  res.body[0]))
        .subscribe(section => {
          if (section.confParamValue === '') {
              const modal = this.modalService.open(EmailSettingNotificationComponent, {centered: true});
              modal.componentInstance.section = section;

              modal.componentInstance.action.subscribe(action =>
                this.router.navigate(['/app-management/settings/application-config']));
          }
        });
  }
}
