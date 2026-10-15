import {ChangeDetectorRef, Component, OnDestroy, OnInit, ViewChild} from '@angular/core';
import {HttpErrorResponse} from '@angular/common/http';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {Subscription} from 'rxjs';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {SectionConfigParamType} from '../../../shared/types/configuration/section-config-param.type';
import {TfaDisableModalComponent} from '../../components/tfa-disable-modal/tfa-disable-modal.component';
import {UserResponse} from '../../domain/tfa.model';
import {FederationTfaService} from '../../services/federation-tfa-service';

import {UtmTfaConfCheckComponent} from '../../../shared/components/utm/config/shared/components/utm-tfa-conf-check/utm-tfa-conf-check.component';
import {UtmToggleComponent} from '../../../shared/components/utm/util/utm-toggle/utm-toggle.component';

@Component({
  selector: 'app-federation-tfa-config-page',
  templateUrl: './federation-tfa-config.page.component.html'
})
export class FederationTfaConfigPageComponent implements OnInit, OnDestroy {
  @ViewChild('tfaCheck') tfaCheck: UtmTfaConfCheckComponent;
  @ViewChild('toggle') toggle: UtmToggleComponent;
  user: UserResponse;
  switchActive = false;
  config: SectionConfigParamType[] = [];
  tfaChecked = false;
  saving = false;
  private sub: Subscription;

  constructor(private tfaService: FederationTfaService,
              private toast: UtmToastService,
              private modalService: NgbModal,
              private cdr: ChangeDetectorRef) {}

  get enabled(): boolean {
    return this.user ? this.user.tfaEnabled : false;
  }

  ngOnInit(): void {
    this.loadUser();
  }

  loadUser(): void {
    this.sub = this.tfaService.me().subscribe(res => {
      this.user = res.body;
      this.switchActive = this.enabled;
      const method = (this.user && this.user.tfaMethod) ? this.user.tfaMethod : 'TOTP';
      // SAFETY: UtmTfaConfCheck only reads confParamShort/confParamValue from this array; full param shape not needed here.
      this.config = [
        {confParamShort: 'utmstack.tfa.enable', confParamValue: String(this.enabled)},
        {confParamShort: 'utmstack.tfa.method', confParamValue: method}
      ] as unknown as SectionConfigParamType[];
    });
  }

  onToggle(value: boolean): void {
    if (value) {
      // enable: user must click the verify button and confirm the code
      this.tfaChecked = false;
      this.tfaCheck.initTfa()
    } else {
      // disable: ask password first; switch only flips if API succeeds
      if (this.saving) {
        return;
      }
      this.modalService.open(TfaDisableModalComponent, {centered: true, size: 'sm'}).result.then(
        (password: string) => this.disableTfa(password),
        () => this.rejectSwitch()
      ).catch(() => this.rejectSwitch());
    }
  }

  checkedTFAConfig(value: boolean): void {
    this.tfaChecked = value;
    if (value) {
      this.complete();
    }
  }

  /** Flips the switch back on when the user cancels the password modal. */
  rejectSwitch(): void {
    this.switchActive = true;
    if (this.toggle) {
      this.toggle.active = true;
    }
    this.cdr.detectChanges();
  }

  disableTfa(password: string): void {
    this.saving = true;
    this.tfaService.disableTfa(password).subscribe(() => {
      this.saving = false;
      this.switchActive = false;
      this.loadUser();
      this.toast.showSuccessBottom('Two-factor authentication disabled');
    }, (error: HttpErrorResponse) => {
      this.saving = false;
      this.toast.showError('Error', this.extractError(error));
    });
  }

  private complete(): void {
    this.saving = true;
    this.tfaService.completeConfiguration().subscribe(() => {
      this.saving = false;
      this.switchActive = true;
      this.loadUser();
      this.toast.showSuccessBottom('Two-factor authentication enabled');
    }, (error: HttpErrorResponse) => {
      this.saving = false;
      this.toast.showError('Error', this.extractError(error));
    });
  }

  private extractError(error: HttpErrorResponse): string {
    if (error && error.headers && error.headers.get('X-UtmStack-error')) {
      return error.headers.get('X-UtmStack-error');
    }
    return (error && error.message) ? error.message : 'Error saving two-factor authentication configuration';
  }

  ngOnDestroy(): void {
    if (this.sub) {
      this.sub.unsubscribe();
    }
  }
}
