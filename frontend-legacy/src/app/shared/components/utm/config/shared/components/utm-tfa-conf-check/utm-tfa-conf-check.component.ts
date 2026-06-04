import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {DomSanitizer} from '@angular/platform-browser';
import {AccountService} from '../../../../../../../core/auth/account.service';
import {ModalService} from '../../../../../../../core/modal/modal.service';
import {UtmToastService} from '../../../../../../alert/utm-toast.service';
import {TfaInitResponse, TfaService} from '../../../../../../services/tfa/tfa.service';
import {SectionConfigParamType} from '../../../../../../types/configuration/section-config-param.type';
import {UtmTfaVerificationComponent} from '../../../../../utm-tfa-verification/utm-tfa-verification.component';

@Component({
  selector: 'app-utm-tfa-conf-check',
  templateUrl: './utm-tfa-conf-check.component.html',
  styleUrls: ['./utm-tfa-conf-check.component.css']
})
export class UtmTfaConfCheckComponent implements OnInit {
  @Input() validConfig: boolean;
  @Input() config: SectionConfigParamType[] = [];
  @Input() configToSave: SectionConfigParamType[] = [];
  @Output() isChecked = new EventEmitter<boolean>();
  checking: any;
  email: string;

  constructor(private accountService: AccountService,
              private utmToastService: UtmToastService,
              private tfaService: TfaService,
              private modalService: ModalService,
              private sanitizer: DomSanitizer) {
  }

  ngOnInit() {
  }

  initTfa() {
    const tfaMethod = this.config.find(conf => conf.confParamShort === 'utmstack.tfa.method');
    this.checking = true;
    this.accountService.identity().then(account => {
      this.tfaService.initTfa({
        method: tfaMethod.confParamValue
      }).subscribe((response) => {
        this.checking = false;
        this.openModal(response);
      }, (error) => {
        this.checking = false;
        if (error.status === 400) {
          this.utmToastService.showError('Error initializing TFA',
            'An error occurred while initializing two-factor authentication, please check configuration and try again');
        } else {
          this.utmToastService.showError('Error initializing TFA',
            'An error occurred while initializing two-factor authentication, please contact with the support team');
        }
        this.isChecked.next(false);
      });
    });
  }

  openModal(response: TfaInitResponse) {
    const modalSource = this.modalService.open(UtmTfaVerificationComponent, {
      centered: true,
      size: 'sm'
    });

    modalSource.componentInstance.method = this.config.find(conf => conf.confParamShort === 'utmstack.tfa.method').confParamValue;
    modalSource.componentInstance.qrCodeUrl = response.delivery.target ?
      this.sanitizer.bypassSecurityTrustUrl(`data:image/png;base64,${response.delivery.target}`) : null;
    modalSource.componentInstance.expiresInSeconds = response.expiresInSeconds;

    modalSource.result.then((result) => {
      if (result) {
        console.log('TFA verification successful');
        this.isChecked.next(true);
      }
    });
  }
}
