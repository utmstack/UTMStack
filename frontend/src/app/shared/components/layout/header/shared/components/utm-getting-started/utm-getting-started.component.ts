import {Component, EventEmitter, Input, OnDestroy, OnInit, Output} from '@angular/core';
import {NavigationEnd, Router} from '@angular/router';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {Subscription} from 'rxjs';
import {filter} from 'rxjs/operators';
import {UtmToastService} from '../../../../../../alert/utm-toast.service';
import {GettingStartedBehavior} from '../../../../../../behaviors/getting-started.behavior';
import {GettingStartedService} from '../../../../../../services/getting-started/getting-started.service';
import {ApplicationConfigSectionEnum} from '../../../../../../types/configuration/section-config.type';
import {
  GettingStartedStepDocURLEnum,
  GettingStartedStepEnum,
  GettingStartedStepIconPathEnum,
  GettingStartedStepNameEnum,
  GettingStartedStepUrlEnum,
  GettingStartedStepVideoPathEnum,
  GettingStartedType
} from '../../../../../../types/getting-started/getting-started.type';
import {isSubdomainOfUtmstack} from '../../../../../../util/url.util';


@Component({
  selector: 'app-utm-getting-started',
  templateUrl: './utm-getting-started.component.html',
  styleUrls: ['./utm-getting-started.component.scss']
})
export class UtmGettingStartedComponent implements OnInit, OnDestroy {
  steps: GettingStartedType[] = [];
  stepsPoint: boolean[];
  private routeSub: Subscription;

  constructor(private utmGettingStartedService: GettingStartedService,
              private router: Router,
              private modalService: NgbModal,
              private gettingStartedBehavior: GettingStartedBehavior) {
  }

  ngOnInit() {
    this.getSteps();
    this.gettingStartedBehavior.$init.subscribe((value) => {
      if (value) {
        this.getSteps();
      }
    });
    if (this.router.url.includes(GettingStartedStepUrlEnum.APPLICATION_SETTINGS)
      && !this.isStepCompleted(GettingStartedStepEnum.APPLICATION_SETTINGS)) {
      this.openModal(GettingStartedStepEnum.APPLICATION_SETTINGS);
    }
    this.routeSub = this.router.events.pipe(
      filter(event => event instanceof NavigationEnd)
    ).subscribe((event: NavigationEnd) => {
      const matchedEnumKey = Object.keys(GettingStartedStepUrlEnum)
        .find(key => event.url.includes(GettingStartedStepUrlEnum[key]));

      if (matchedEnumKey) {
        const step = this.steps.find(s => s.stepShort === matchedEnumKey);
        if (step && !step.completed) {
          this.openModal(GettingStartedStepEnum[matchedEnumKey]);
        }
      }
    });
  }

  ngOnDestroy(): void {
    if (this.routeSub) {
      this.routeSub.unsubscribe();
    }
  }

  openModal(step: GettingStartedStepEnum) {
    const modalRef = this.modalService.open(GettingStartedModalComponent,
      {centered: true, size: 'lg', backdrop: 'static', keyboard: false});
    modalRef.componentInstance.step = step;
    modalRef.componentInstance.completeStep.subscribe(() => {
      this.getSteps();
    });
  }

  getSteps() {
    this.utmGettingStartedService.getSteps({page: 0, size: 15}).subscribe(response => {
      this.steps = response.body.sort((a, b) => a.stepOrder - b.stepOrder).slice();
      this.stepsPoint = response.body.sort((a, b) => a.stepOrder - b.stepOrder).map(value => value.completed);
    });
  }

  get countCompleted(): number {
    return this.steps.filter(value => value.completed === true).length;
  }

  get completed(): boolean {
    return this.steps.filter(value => value.completed === true).length === this.steps.length;
  }

  show(): boolean {
    return (!this.steps || this.steps.length === 0) ? false : this.countCompleted < this.steps.length;
  }


  calcPercent(): number {
    return Math.floor((this.countCompleted / this.steps.length) * 100);
  }

  get pointWidth() {
    // Assuming each point has a 1px margin-right except for the last point
    const totalMargin = (this.steps.length - 1);
    const containerWidth = 374; // Adjust this as per your requirement
    return (containerWidth - totalMargin) / this.steps.length;
  }

  getName(step: GettingStartedStepEnum) {
    return GettingStartedStepNameEnum[step];
  }

  isStepCompleted(step: GettingStartedStepEnum): boolean {
    if (!this.steps || this.steps.length === 0) {
      return false;
    }
    return this.steps.filter(value => value.stepShort === step)[0].completed;
  }

  goToGuide(step: GettingStartedStepEnum) {
    const route = GettingStartedStepUrlEnum[step];
    let queryParams = {};

    switch (step) {
      case GettingStartedStepEnum.DASHBOARD_BUILDER:
        queryParams = {mode: 'edit', dashboardId: 7, dashboardName: 'threat_activity'};
        break;

      case GettingStartedStepEnum.THREAT_MANAGEMENT:
        queryParams = {alertType: 'ALERT'};
        break;

      case GettingStartedStepEnum.APPLICATION_SETTINGS:
        queryParams = {sections: JSON.stringify([ApplicationConfigSectionEnum.EMAIL, ApplicationConfigSectionEnum.ALERTS])};
        break;

      default:
        break;
    }

    this.router.navigate([route], {
      queryParams
    });
  }

  isLastToBeCompleted(stepEnum: GettingStartedStepEnum): boolean {
    // Sort steps based on stepOrder
    const sortedSteps = this.steps.slice().reverse();
    // Check if the given stepEnum is the first in the list of uncompleted steps
    return sortedSteps[0].stepShort === stepEnum && sortedSteps[0].completed;
  }
}


@Component({
  selector: 'app-modal-getting-started',
  template: `
    <app-utm-modal-header [name]="name" *ngIf="name" [showCloseButton]="false"></app-utm-modal-header>
    <div class="modal-body" *ngIf="step">
      <ng-container [ngSwitch]="step" class="">
        <h5 class="text-justify mb-3 font-weight-thin" *ngSwitchCase="gettingStartedStepEnum.THREAT_MANAGEMENT">
          You are now seeing the alert management section. Here are all the alerts created by the system.
          Now change the status of one of the sample alerts to "Complete", and click on the "Integrations" Menu
        </h5>
        <h5 class="text-justify mb-3 font-weight-thin" *ngSwitchCase="gettingStartedStepEnum.DASHBOARD_BUILDER">
          UTMStack dashboards are interactive resources that can be easily created and modified. To use a dashboard, click any of
          the visualizations.
        </h5>
        <h5 class="text-justify mb-3 font-weight-thin" *ngSwitchCase="gettingStartedStepEnum.INTEGRATIONS">
          Configure your first integration in less than five minutes by clicking "View Integration".
          We recommend getting started with the Windows or Linux Agent.
        </h5>
      </ng-container>
      <img height="480" controls *ngIf="videoPath" style="width: 100%" [src]="videoPath | safe:'resourceUrl'" [alt]="name">

    </div>
    <div [ngClass]="{'justify-content-between': urlDoc}"  class="modal-footer d-flex  align-items-center">
      <app-utm-online-documentation *ngIf="urlDoc" [path]="urlDoc" text="Learn more"></app-utm-online-documentation>
      <button class="btn utm-button utm-button-primary d-flex justify-content-start"
              (click)="complete()">Continue
      </button>
    </div>
  `,
})
export class GettingStartedModalComponent implements OnInit {
  @Input() step: GettingStartedStepEnum;
  gettingStartedStepEnum = GettingStartedStepEnum;
  @Output() completeStep = new EventEmitter<GettingStartedStepEnum>();
  urlDoc: string;
  videoPath: string;
  icon: string;
  name: string;

  constructor(public activeModal: NgbActiveModal,
              private utmToastService: UtmToastService,
              private utmGettingStartedService: GettingStartedService) {
  }

  ngOnInit() {
    const  inSaas = isSubdomainOfUtmstack();
    this.icon = GettingStartedStepIconPathEnum[this.step];
    this.name = GettingStartedStepNameEnum[this.step];

    if (this.step === GettingStartedStepEnum.APPLICATION_SETTINGS && inSaas) {
      this.videoPath = GettingStartedStepVideoPathEnum.APPLICATION_SAAS_SETTINGS;
    } else {
      this.videoPath = GettingStartedStepVideoPathEnum[this.step];
    }

    this.urlDoc = (this.step === GettingStartedStepEnum.DASHBOARD_BUILDER || this.step === GettingStartedStepEnum.THREAT_MANAGEMENT)
      ? GettingStartedStepDocURLEnum[this.step] : null;
  }

  complete() {
    this.utmGettingStartedService.completeStep(this.step.toString()).subscribe(value => {
      this.completeStep.emit(this.step);
      this.activeModal.close();
      this.utmToastService.showSuccessBottom('Step completed');
    });
  }
}

@Component({
  selector: 'app-modal-getting-started-complete',
  template: `
    <app-utm-modal-header [name]="'Getting started completed'" [showCloseButton]="false"></app-utm-modal-header>
    <div class="modal-body">
      <div class="d-flex justify-content-center align-items-center">
        <span inlineSVG="/assets/icons/system/SPEED_OPTIMIZATION.svg" class="svg-icon svg-icon-grey svg-icon-10x"></span>
      </div>
      <p class="text-justify font-size-lg mt-3">
        You have completed the initial instance setup, if you have more question, visit the
        <a class="font-size-base text-blue-800"
           href="https://docs.utmstack.com"
           target="_blank"> online documentation</a>
      </p>
    </div>
    <div class="modal-footer">
      <button class="btn utm-button utm-button-primary"
              (click)="exit()">Close
      </button>
    </div>
  `,
})
export class GettingStartedFinishedModalComponent implements OnInit {
  constructor(public activeModal: NgbActiveModal) {
  }

  ngOnInit() {
  }

  exit() {
    this.activeModal.close();
  }
}
