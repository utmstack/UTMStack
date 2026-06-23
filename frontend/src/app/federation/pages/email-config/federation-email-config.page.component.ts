import {Component} from '@angular/core';
import {
  FEDERATION_EMAIL_CONFIG_PARAMS,
  FEDERATION_EMAIL_CONFIG_SECTION
} from './federation-email-config-params.const';

@Component({
  selector: 'app-federation-email-config-page',
  templateUrl: './federation-email-config.page.component.html',
  styleUrls: ['./federation-email-config.page.component.scss']
})
export class FederationEmailConfigPageComponent {
  readonly section = FEDERATION_EMAIL_CONFIG_SECTION;
  readonly params = FEDERATION_EMAIL_CONFIG_PARAMS;
}
