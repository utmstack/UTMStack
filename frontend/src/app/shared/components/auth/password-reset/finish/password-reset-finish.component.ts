import {AfterViewInit, Component, ElementRef, OnInit, Renderer} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {NgbModalRef} from '@ng-bootstrap/ng-bootstrap';
import {ModalService} from '../../../../../core/modal/modal.service';
import {PasswordResetFinishService} from './password-reset-finish.service';


@Component({
  selector: 'app-password-reset-finish',
  templateUrl: './password-reset-finish.component.html',
  styleUrls: ['./password-reset-finish-component.scss']
})
export class PasswordResetFinishComponent implements OnInit, AfterViewInit {
  confirmPassword: string;
  doNotMatch: string;
  error: string;
  keyMissing: boolean;
  resetAccount: any;
  success: string;
  modalRef: NgbModalRef;
  key: string;
  sending = false;

  constructor(
    private passwordResetFinishService: PasswordResetFinishService,
    private loginModalService: ModalService,
    private route: ActivatedRoute,
    private router: Router,
    private elementRef: ElementRef,
    private renderer: Renderer
  ) {
  }

  ngOnInit() {
    this.route.queryParams.subscribe(params => {
      this.key = params.key;
    });
    this.resetAccount = {};
    this.keyMissing = !this.key;
  }

  ngAfterViewInit() {
    if (this.elementRef.nativeElement.querySelector('#password') != null) {
      this.renderer.invokeElementMethod(this.elementRef.nativeElement.querySelector('#password'), 'focus', []);
    }
  }

  finishReset() {
    this.doNotMatch = null;
    this.error = null;
    if (this.resetAccount.password !== this.confirmPassword) {
      this.doNotMatch = 'ERROR';
    } else {
      this.sending = true;
      this.passwordResetFinishService
        .save({key: this.key, newPassword: this.resetAccount.password}).subscribe(
        () => {
          this.success = 'OK';
          this.sending = false;
        },
        () => {
          this.success = null;
          this.error = 'ERROR';
          this.sending = false;
        }
      );
    }
  }

  login() {
    // this.activeModal.close('close');
    this.router.navigate(['']);
  }
}
