import {Component, OnDestroy, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {fromEvent, Subject} from 'rxjs';
import {takeUntil} from 'rxjs/operators';
import {GettingStartedBehavior} from 'src/app/shared/behaviors/getting-started.behavior';
import {AccountService} from '../../../../core/auth/account.service';
import {ONLINE_DOCUMENTATION_BASE} from '../../../constants/global.constant';
import {GettingStartedService} from '../../../services/getting-started/getting-started.service';
import {ApplicationConfigSectionEnum} from "../../../types/configuration/section-config.type";
import {GettingStartedStepEnum} from '../../../types/getting-started/getting-started.type';
import {isSubdomainOfUtmstack} from '../../../util/url.util';
import {UtmAdminChangeEmailComponent} from '../utm-admin-change-email/utm-admin-change-email.component';

@Component({
  selector: 'app-welcome-to-utmstack',
  templateUrl: './welcome-to-utmstack.component.html',
  styleUrls: ['./welcome-to-utmstack.component.scss']
})
export class WelcomeToUtmstackComponent implements OnInit, OnDestroy {
  private unsubscriber: Subject<void> = new Subject<void>();
  accountSetup = true;
  onlineDoc = ONLINE_DOCUMENTATION_BASE;
  inSass: boolean;

  ngOnDestroy(): void {
    this.unsubscriber.next();
    this.unsubscriber.complete();
  }

  constructor(private router: Router,
              private accountService: AccountService,
              private utmGettingStartedService: GettingStartedService,
              private gettingStartedBehavior: GettingStartedBehavior,
              private modalService: NgbModal) {

  }

  ngOnInit(): void {
    this.inSass = isSubdomainOfUtmstack();
    history.pushState(null, '');

    fromEvent(window, 'popstate').pipe(
      takeUntil(this.unsubscriber)
    ).subscribe((_) => {
      history.pushState(null, '');
    });
  }

  exit() {
    this.setUpAdmin(false);
  }

  setUpAdmin(gettingStarted: boolean) {
    this.accountService.identity().then(account => {
      const modal = this.modalService.open(UtmAdminChangeEmailComponent, {
        centered: true,
        size: 'sm',
        backdrop: 'static',
        windowClass: 'adminSetupModal',
        keyboard: false,
      });
      modal.componentInstance.account = account;
      modal.componentInstance.setupSuccess.subscribe(() => {
        if (gettingStarted) {
          this.utmGettingStartedService.completeStep(GettingStartedStepEnum.SET_ADMIN_USER).subscribe(() => {
            this.router.navigate(['/app-management/settings/application-config'], {
             queryParams: {sections: JSON.stringify(this.sectionsOrder())}
           });
          });
        } else {
          this.router.navigate(['/integrations/explore']);
        }
        this.accountSetup = true;
      });
    });
  }

  gettingStarted() {
    this.utmGettingStartedService.initialize(this.inSass).subscribe(value => {
      this.gettingStartedBehavior.$init.next(true);
      this.setUpAdmin(true);
    });
  }

  sectionsOrder() {
    if (this.inSass) {
      return [ApplicationConfigSectionEnum.ALERTS,
        ApplicationConfigSectionEnum.EMAIL,
        ApplicationConfigSectionEnum.TFA,
        ApplicationConfigSectionEnum.DATE_SETTINGS];
    } else {
      return [ApplicationConfigSectionEnum.EMAIL,
        ApplicationConfigSectionEnum.ALERTS,
        ApplicationConfigSectionEnum.TFA,
        ApplicationConfigSectionEnum.DATE_SETTINGS];
    }
  }
}
