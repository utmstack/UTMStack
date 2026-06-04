import {HttpErrorResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Router} from '@angular/router';
import {TranslateService} from '@ngx-translate/core';
import {ToastrManager} from 'ng6-toastr-notifications';

@Injectable({providedIn: 'root'})
export class UtmToastService {
  constructor(public toastr: ToastrManager,
              private translateService: TranslateService,
              private router: Router) {
  }

  showSuccess(msg) {
    this.toastr.successToastr(msg, 'Logged', 'Success login');
  }

  showError(title, msg) {
    this.toastr.errorToastr(msg, title, {position: 'bottom-right'});
  }

  showWarning(msg, title) {
    this.toastr.warningToastr(msg, title, {position: 'bottom-right'});
  }

  showInfoAssets(msg, title) {
    this.toastr.infoToastr(msg, title, {position: 'bottom-right'});
  }

  showInfo(title: string, message: string) {
    this.toastr.infoToastr(message, title, {position: 'bottom-right'});
  }

  showCustom(msg) {
    this.toastr.customToastr(
      '<span class="custom-toast" style=\'color: green; font-size: 12px; text-align: center;\'>' + msg + '</span>',
      null,
      {enableHTML: true, position: 'bottom-right'}
    );
  }

  showToast(position: any = 'top-left') {
    // this.toastr.infoToastr('This is a toast.', 'Toast', {position: position});
  }

  showSuccessProcess(title: string, msg: string) {
    this.toastr.successToastr(this.extractMessage(msg), this.extractTitle(title));
  }

  showErrorResponse(title: string, error: HttpErrorResponse) {
    this.toastr.errorToastr(this.buildMessage(error.error.code), this.extractTitle(title));
  }

  showSuccessBottom(msg) {
    this.toastr.successToastr(msg, 'Success!',
      {position: 'bottom-right', toastTimeout: 2000});
  }

  buildMessage(code: number): string {
    return '';
  }

  extractMessage(key: string): string {
    let msg = '';
    this.translateService.get(key).subscribe(value => {
      msg = value;
    });
    return msg;
  }


  private extractTitle(title: string): string {
    let tit = '';
    this.translateService.get(title).subscribe(value => {
      tit = value;
    });
    return tit;
  }
}
