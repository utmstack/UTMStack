import {Component, OnInit} from '@angular/core';
import {UtmConfigParamsService} from '../../../shared/services/config/utm-config-params.service';
import {SectionConfigParamType} from '../../../shared/types/configuration/section-config-param.type';
import {
  FEDERATION_EMAIL_CONFIG_PARAMS,
  FEDERATION_EMAIL_CONFIG_SECTION
} from './federation-email-config-params.const';

@Component({
  selector: 'app-federation-email-config-page',
  templateUrl: './federation-email-config.page.component.html',
  styleUrls: ['./federation-email-config.page.component.scss']
})
export class FederationEmailConfigPageComponent implements OnInit {
  readonly section = FEDERATION_EMAIL_CONFIG_SECTION;
  readonly params: SectionConfigParamType[] = FEDERATION_EMAIL_CONFIG_PARAMS.map(p => ({...p}));

  constructor(private paramsService: UtmConfigParamsService) {}

  ngOnInit(): void {
    this.paramsService.query({
      page: 0,
      size: 10000,
      'sectionId.equals': this.section.id,
      sort: 'id,asc'
    }).subscribe(response => {
      const byShort = new Map<string, SectionConfigParamType>();
      (response.body || []).forEach(p => byShort.set(p.confParamShort, p));
      this.params.forEach(p => {
        const server = byShort.get(p.confParamShort);
        if (server) {
          p.id = server.id;
          p.confParamValue = server.confParamValue;
        }
      });
      this.maskPasswordIfPreviouslySet();
    });
  }

  private maskPasswordIfPreviouslySet(): void {
    const password = this.params.find(p => p.confParamShort === 'utmstack.mail.password');
    if (!password || !this.isEmpty(password.confParamValue)) {
      return;
    }
    const others = this.params.filter(p => p !== password);
    const allOthersFilled = others.length > 0 && others.every(p => !this.isEmpty(p.confParamValue));
    if (allOthersFilled) {
      password.confParamPlaceholder = '••••••••';
      password.confParamRequired = false;
    }
  }

  private isEmpty(value: unknown): boolean {
    return value === null || value === undefined || value === '';
  }
}
