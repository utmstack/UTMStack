import { HttpErrorResponse } from '@angular/common/http';
import {AfterViewInit, Component, ElementRef, OnInit} from '@angular/core';
import {NgbModalRef} from '@ng-bootstrap/ng-bootstrap';
import {ModalService} from '../../core/modal/modal.service';
import {LoginComponent} from '../../shared/components/auth/login/login.component';
import {EMAIL_ALREADY_USED_TYPE, LOGIN_ALREADY_USED_TYPE} from '../../shared/constants/error.constants';


import {Register} from './register.service';

@Component({
    standalone: false,
    selector: 'app-register',
    templateUrl: './register.component.html',
})
export class RegisterComponent implements OnInit, AfterViewInit {
  confirmPassword: string;
  doNotMatch: string;
  error: string;
  errorEmailExists: string;
  errorUserExists: string;
  registerAccount: any;
  success: boolean;
  modalRef: NgbModalRef;

  constructor(
    private loginModalService: ModalService,
    private registerService: Register,
    private elementRef: ElementRef
  ) {
  }

  ngOnInit() {
    this.success = false;
    this.registerAccount = {};
  }

  ngAfterViewInit() {
    const el = this.elementRef.nativeElement.querySelector('#login') as HTMLInputElement;
    if (el) { el.focus(); }
  }

  register() {
    if (this.registerAccount.password !== this.confirmPassword) {
      this.doNotMatch = 'ERROR';
    } else {
      this.doNotMatch = null;
      this.error = null;
      this.errorUserExists = null;
      this.errorEmailExists = null;
      this.registerAccount.langKey = 'en';
      this.registerService.save(this.registerAccount).subscribe(
        () => {
          this.success = true;
        },
        response => this.processError(response)
      );
    }
  }

  openLogin() {
    this.modalRef = this.loginModalService.open(LoginComponent, undefined, '');
  }

  private processError(response: HttpErrorResponse) {
    this.success = null;
    if (response.status === 400 && response.error.type === LOGIN_ALREADY_USED_TYPE) {
      this.errorUserExists = 'ERROR';
    } else if (response.status === 400 && response.error.type === EMAIL_ALREADY_USED_TYPE) {
      this.errorEmailExists = 'ERROR';
    } else {
      this.error = 'ERROR';
    }
  }
}
