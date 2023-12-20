import {Directive, Input, TemplateRef, ViewContainerRef} from '@angular/core';
import {LicenceChangeBehavior} from '../../behaviors/licence-change.behavior';
import {UtmClientService} from '../../services/license/utm-client.service';

@Directive({
  selector: '[appEnterpriseLicence]'
})
export class HasEnterpriseLicenseDirective {
  private licenseType: string;

  constructor(
    private utmClientService: UtmClientService,
    private templateRef: TemplateRef<any>,
    private licenceChangeBehavior: LicenceChangeBehavior,
    private viewContainerRef: ViewContainerRef
  ) {
  }

  @Input()
  set appEnterpriseLicence(value: string | string[]) {
    this.updateView();
    // Get notified each time authentication state changes.
    this.licenceChangeBehavior.$licenceChange.subscribe(licenceChange => {
      if (licenceChange) {
        this.updateView();
      }
    });
  }

  private updateView(): void {
    this.utmClientService.query().subscribe(response => {
      this.viewContainerRef.clear();
      if (response.body[0].clientLicenceVerified) {
        this.viewContainerRef.createEmbeddedView(this.templateRef);
      }
    });
  }
}
