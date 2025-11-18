import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {ClientAuthMethod, ProviderType, UtmIdentityProvider, } from '../../models/utm-identity-provider.model';

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
  editMode = false;

  providerForm!: FormGroup;

  providerTypes = Object.values(ProviderType).map((value) => ({
    label: value,
    value
  }));

  availableScopes = [
    'openid',
    'email',
    'profile',
  ];

  constructor(private fb: FormBuilder) {}

  ngOnInit(): void {
    this.editMode = !!this.provider;
    this.initForm();
    if (this.editMode) {

      this.providerTypes = this.providerTypes.filter(pt => pt.value === this.provider.providerType);
      this.providerForm.patchValue(this.provider);

      this.providerForm.get('scopes').setValue(this.provider.scopes?.split(',') || []);
      this.providerForm.get('allowedDomains').setValue(this.provider.allowedDomains?.split(',') || []);
    }
  }

  initForm(): void {
    this.providerForm = this.fb.group({
      name: ['', Validators.required],
      providerType: ['', Validators.required],
      clientId: ['', Validators.required],
      clientSecret: this.editMode ? [''] : ['', Validators.required],
      redirectUri: ['', Validators.required],
      authUri: ['', Validators.required],
      tokenUri: ['', Validators.required],
      userInfoUri: ['', Validators.required],
      jwksUri: [''],
      scopes: ['', Validators.required],
      allowedDomains: [''],
      active: [true]
    });
  }

  onProviderTypeChange(): void {
    const type = this.providerForm.get('providerType').value;
    const presets = this.getProviderPresets(type.value);

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
