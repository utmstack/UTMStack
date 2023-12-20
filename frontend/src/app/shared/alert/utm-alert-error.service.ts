import {Injectable} from '@angular/core';
import {TranslateService} from '@ngx-translate/core';
import {Subscription} from 'rxjs';
import {UtmToastService} from './utm-toast.service';

@Injectable()
export class UtmAlertErrorService {
  cleanHttpErrorListener: Subscription;

  /* tslint:disable */
  constructor(private translateService: TranslateService,
              public alertService: UtmToastService
  ) {
    /* tslint:enable */
    // this.cleanHttpErrorListener = eventManager.subscribe('inverdiamond.httpError', response => {
    //   const httpErrorResponse = response.content;
    //   console.log(response);
    //   switch (httpErrorResponse.status) {
    //     // connection refused, server not reachable
    //     case 0:
    //       // this.addErrorAlert('Server not reachable', 'error.server.not.reachable');
    //       break;
    //
    //     case  400:
    //       // this.addErrorAlert(response.error, 'error.server.not.reachable');
    //       break;
    //
    //     case 404:
    //       // this.addErrorAlert('Not found', 'error.url.not.found');
    //       break;
    //   }
    // });
  }

  addErrorAlert(message, key?, data?) {
    key = key && key !== null ? key : message;
    this.alertService.showError('Sorry', message);
  }
}
