import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {ClientAuthMethod, ProviderType, UtmIdentityProvider,} from '../../models/utm-identity-provider.model';

@Component({
  selector: 'app-provider-form',
  templateUrl: './provider-form.component.html',
  styleUrls: ['./provider-form.component.scss']
})
export class ProviderFormComponent implements OnInit {
  @Input() provider?: UtmIdentityProvider;
  @Input() loading = false;
  @Input() testingConnection = false;

  @Output() save = new EventEmitter<UtmIdentityProvider>();
  @Output() test = new EventEmitter<void>();
  @Output() cancel = new EventEmitter<void>();

  providerForm!: FormGroup;

  providerTypes = Object.values(ProviderType);
  authMethods = Object.values(ClientAuthMethod);

  constructor(private fb: FormBuilder) {}

  ngOnInit(): void {
    this.initForm();
    if (this.provider) {
      this.providerForm.patchValue(this.provider);
    }
  }

  initForm(): void {
    this.providerForm = this.fb.group({
      name: ['', Validators.required],
      providerType: ['Google', Validators.required],
      clientId: ['', Validators.required],
      clientSecret: ['', Validators.required],
      redirectUri: ['', Validators.required],
      clientAuthMethod: ['client_secret_basic'],
      authUri: ['', Validators.required],
      tokenUri: ['', Validators.required],
      userInfoUri: ['', Validators.required],
      jwksUri: [''],
      scopes: ['openid,email,profile', Validators.required],
      allowedDomains: [''],
      active: [true]
    });
  }

  onProviderTypeChange(): void {
    const type = this.providerForm.get('providerType').value;
    const presets = this.getProviderPresets(type);

    if (presets) {
      this.providerForm.patchValue(presets);
    }
  }

  getProviderPresets(type: string): Partial<UtmIdentityProvider> | null {

    const presets: Record<string, Partial<UtmIdentityProvider>> = {
      [ProviderType.GOOGLE]: {
        authUri: 'https://accounts.google.com/o/oauth2/auth',
        tokenUri: 'https://oauth2.googleapis.com/token',
        userInfoUri: 'https://www.googleapis.com/oauth2/v2/userinfo',
        jwksUri: 'https://www.googleapis.com/oauth2/v3/certs',
        scopes: 'openid,email,profile'
      },
      [ProviderType.MICROSOFT]: {
        authUri: 'https://login.microsoftonline.com/common/oauth2/v2.0/authorize',
        tokenUri: 'https://login.microsoftonline.com/common/oauth2/v2.0/token',
        userInfoUri: 'https://graph.microsoft.com/v1.0/me',
        scopes: 'openid,email,profile,User.Read'
      }
    };

    return presets[type] || null;
  }
}
