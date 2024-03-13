import {Component, Input, OnInit} from '@angular/core';
import {AccountService} from '../../../../../core/auth/account.service';
import {UtmToastService} from '../../../../alert/utm-toast.service';
import {UtmConfigEmailCheckService} from '../../../../services/config/utm-config-email-check.service';
import {SectionConfigParamType} from "../../../../types/configuration/section-config-param.type";

@Component({
  selector: 'app-utm-email-conf-check',
  templateUrl: './utm-email-conf-check.component.html',
  styleUrls: ['./utm-email-conf-check.component.css']
})
export class UtmEmailConfCheckComponent implements OnInit {
  @Input() validConfig: boolean;
  @Input() emailConfig: SectionConfigParamType[] = [];
  checking: any;
  email: string;

  constructor(private utmConfigEmailCheckService: UtmConfigEmailCheckService,
              private accountService: AccountService,
              private utmToastService: UtmToastService) {
  }

  ngOnInit() {
  }

  testConfig() {
    console.log(this.emailConfig);
    this.checking = true;
    this.accountService.identity().then(account => {
      this.utmConfigEmailCheckService.checkEmail().subscribe(() => {
        this.checking = false;
        this.utmToastService.showSuccessBottom('A test email has been sent to ' + account.email + '; please check your inbox.');
      }, (error) => {
        this.checking = false;
        if (error.status === 400) {
          this.utmToastService.showError('Error sending email',
            'An error occurred while sending the test email, please check configuration and try again');
        } else {
          this.utmToastService.showError('Error sending email',
            'An error occurred while sending the test email, please contact with the support team');
        }
      });
    });
  }
}
