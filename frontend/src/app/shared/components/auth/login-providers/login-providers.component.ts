import { Component, OnInit } from '@angular/core';
import {IdentityProviderDto, LoginProviderService} from '../../../services/login-provider.service';

@Component({
  selector: 'app-login-providers',
  templateUrl: './login-providers.component.html',
  styleUrls: ['./login-providers.component.scss']
})
export class LoginProvidersComponent implements OnInit {

  request = {
    page: 0,
    size: 10
  };

  providers: IdentityProviderDto[] = [] as IdentityProviderDto[];

  constructor(private loginProviderService: LoginProviderService) { }

  ngOnInit() {
    this.loadAllProviders();
  }

  loadAllProviders() {
    this.loginProviderService.getAllProviders(this.request).subscribe(
      response => {
        console.log('Login Providers:', response.body);
        this.providers = response.body || [];
      },
      error => {
        console.error('Error fetching login providers:', error);
        this.providers = [];
      }
    );
  }

  loginWithProvider(provider: IdentityProviderDto) {
    if (!provider.active) {
      console.warn(`Provider ${provider.name} is inactive.`);
      return;
    }

    this.loginProviderService.loginWithProvider(provider.providerType.toLowerCase());
  }
}
