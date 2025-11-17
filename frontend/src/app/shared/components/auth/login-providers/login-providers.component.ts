import { Component, OnInit } from '@angular/core';
import {
  PROVIDER_ICONS,
  ProviderType,
  UtmIdentityProvider
} from '../../../../app-management/identity-provider/shared/models/utm-identity-provider.model';
import {
  LoginProviderService,
} from '../../../services/login-provider.service';

@Component({
  selector: 'app-login-providers',
  templateUrl: './login-providers.component.html',
  styleUrls: ['./login-providers.component.scss']
})
export class LoginProvidersComponent implements OnInit {

  request = {
    'active.equals': false,
    page: 0,
    size: 10
  };

  providers: UtmIdentityProvider[] = [] as UtmIdentityProvider[];

  constructor(private loginProviderService: LoginProviderService) { }

  ngOnInit() {
    this.loadAllProviders();
  }

  loadAllProviders() {
    this.loginProviderService.getProviders(this.request).subscribe(
      response => {
        this.providers = response.body || [];
      },
      error => {
        console.error('Error fetching login providers:', error);
        this.providers = [];
      }
    );
  }

  getProviderIcon(providerType: ProviderType) {
    return PROVIDER_ICONS[providerType] || 'bi-shield-lock';
  }

  loginWithProvider(provider: UtmIdentityProvider) {
    if (!provider.active) {
      console.warn(`Provider ${provider.name} is inactive.`);
      return;
    }

    this.loginProviderService.loginWithProvider(provider.providerType.toLowerCase());
  }
}
