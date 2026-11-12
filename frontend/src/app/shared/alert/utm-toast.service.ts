import { HttpErrorResponse } from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Router} from '@angular/router';
import {TranslateService} from '@ngx-translate/core';
import {ToastrService} from 'ngx-toastr';

/**
 * Toast facade. Backed by ngx-toastr (migrated from ng6-toastr-notifications at
 * the Angular 16 stage — ng6-toastr imports ReflectiveInjector, removed in
 * Angular 16, so it cannot compile at 16+). The public method names are
 * unchanged so the ~94 call sites across the app are untouched.
 *
 * ng6-toastr -> ngx-toastr mapping used here:
 *   successToastr/errorToastr/warningToastr/infoToastr(msg,title,opts)
 *     -> success/error/warning/info(msg,title,opts)
 *   customToastr(msg,null,{enableHTML,position}) -> show(msg,null,{enableHtml,positionClass})
 *   position:'bottom-right'  -> positionClass:'toast-bottom-right'
 *   enableHTML:true          -> enableHtml:true
 *   toastTimeout:N           -> timeOut:N
 *   (no position)            -> top-right (matches both libraries' default)
 */
@Injectable({providedIn: 'root'})
export class UtmToastService {
  constructor(public toastr: ToastrService,
              private translateService: TranslateService,
              private router: Router) {
  }

  showSuccess(msg) {
    this.toastr.success(msg, 'Logged');
  }

  showError(title, msg) {
    this.toastr.error(msg, title, {positionClass: 'toast-bottom-right'});
  }

  showWarning(msg, title) {
    this.toastr.warning(msg, title, {positionClass: 'toast-bottom-right'});
  }

  showInfoAssets(msg, title) {
    this.toastr.info(msg, title, {positionClass: 'toast-bottom-right'});
  }

  showInfo(title: string, message: string) {
    this.toastr.info(message, title, {positionClass: 'toast-bottom-right'});
  }

  showCustom(msg) {
    this.toastr.show(
      '<span class="custom-toast" style=\'color: green; font-size: 12px; text-align: center;\'>' + msg + '</span>',
      null,
      {enableHtml: true, positionClass: 'toast-bottom-right'}
    );
  }

  showToast(position: any = 'top-left') {
    // this.toastr.info('This is a toast.', 'Toast', {position: position});
  }

  showSuccessProcess(title: string, msg: string) {
    this.toastr.success(this.extractMessage(msg), this.extractTitle(title));
  }

  showErrorResponse(title: string, error: HttpErrorResponse) {
    this.toastr.error(this.buildMessage(error.error.code), this.extractTitle(title));
  }

  showSuccessBottom(msg) {
    this.toastr.success(msg, 'Success!',
      {positionClass: 'toast-bottom-right', timeOut: 2000});
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
