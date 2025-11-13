import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {ClientAuthMethod, ProviderType, UtmIdentityProvider} from './shared/models/utm-identity-provider.model';
import {UtmIdentityProviderService} from './shared/services/utm-identity-provider.service';

@Component({
  selector: 'app-identity-provider-config',
  templateUrl: './identity-provider.component.html',
  styleUrls: ['./identity-provider.component.scss']
})
export class IdentityProviderComponent implements OnInit {
  providers: UtmIdentityProvider[] = [];
  providerForm: FormGroup;
  showModal = false;
  editMode = false;
  loading = false;
  testingConnection = false;
  selectedProvider: UtmIdentityProvider | null = null;

  providerTypes = Object.values(ProviderType);
  authMethods = Object.values(ClientAuthMethod);

  providerTemplates: Record<ProviderType, Partial<UtmIdentityProvider>> = {
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

  constructor(
    private fb: FormBuilder,
    private providerService: UtmIdentityProviderService,
    private toast: UtmToastService) {

    this.providerForm = this.createForm();
  }

  ngOnInit(): void {
    this.loadProviders();
  }

  createForm(): FormGroup {
    return this.fb.group({
      name: ['', Validators.required],
      providerType: [ProviderType.GOOGLE, Validators.required],
      clientId: ['', Validators.required],
      clientSecret: ['', Validators.required],
      redirectUri: ['', [Validators.required, Validators.pattern(/^https?:\/\/.+/)]],
      authUri: ['', [Validators.required, Validators.pattern(/^https?:\/\/.+/)]],
      tokenUri: ['', [Validators.required, Validators.pattern(/^https?:\/\/.+/)]],
      userInfoUri: ['', [Validators.required, Validators.pattern(/^https?:\/\/.+/)]],
      jwksUri: ['', Validators.pattern(/^https?:\/\/.+/)],
      clientAuthMethod: [ClientAuthMethod.CLIENT_SECRET_BASIC],
      scopes: ['openid,email,profile', Validators.required],
      allowedDomains: [''],
      active: [true]
    });
  }

  loadProviders(): void {
    this.loading = true;
    this.providerService.query().subscribe({
      next: (res) => {
        this.providers = res.body || [];
        this.loading = false;
      },
      error: () => {
        this.toast.showError('Error', 'An error occurred while loading providers');
        this.loading = false;
      }
    });
  }

  openModal(provider?: UtmIdentityProvider): void {
    this.editMode = !!provider;
    this.selectedProvider = provider || null;

    if (provider) {
      this.providerForm.patchValue(provider);
    } else {
      this.providerForm.reset({
        providerType: ProviderType.GOOGLE,
        active: true,
        ...this.providerTemplates[ProviderType.GOOGLE]
      });
    }

    this.showModal = true;
  }

  closeModal(): void {
    this.showModal = false;
    this.providerForm.reset();
    this.selectedProvider = null;
  }

  onProviderTypeChange(): void {
    const providerType = this.providerForm.get('providerType').value;
    const template = this.providerTemplates[providerType];

    if (template && !this.editMode) {
      this.providerForm.patchValue(template);
    }
  }

  saveProvider(): void {
    if (this.providerForm.invalid) {
      this.toast.showError( 'Error', 'Please fill all required fields correctly');
      return;
    }

    this.loading = true;
    const provider: UtmIdentityProvider = {
      ...this.providerForm.value,
      id: this.selectedProvider.id
    };

    const request = this.editMode
      ? this.providerService.update(provider)
      : this.providerService.create(provider);

    request.subscribe({
      next: () => {
        this.toast.showSuccess(`Provider ${this.editMode ? 'updated' : 'created'} successfully`);
        this.loadProviders();
        this.closeModal();
      },
      error: (err) => {
        this.toast.showError('Error', 'An error occurred while saving the provider');
        this.loading = false;
      }
    });
  }

  deleteProvider(provider: UtmIdentityProvider): void {
    if (!confirm(`Are you sure you want to delete ${provider.name}?`)) {
      return;
    }

    this.providerService.delete(provider.id).subscribe({
      next: () => {
        this.toast.showSuccess('Provider deleted successfully');
        this.loadProviders();
      },
      error: () => {
        this.toast.showError( 'Error', 'An error occurred while deleting the provider');
      }
    });
  }

  toggleActive(provider: UtmIdentityProvider): void {
    this.providerService.toggleActive(provider.id).subscribe({
      next: () => {
        this.toast.showSuccess(`Provider ${provider.active ? 'deactivated' : 'activated'}`);
        this.loadProviders();
      },
      error: () => {
        this.toast.showError( 'Error', 'An error occurred while updating provider status');
      }
    });
  }

  testConnection(): void {
    if (this.providerForm.invalid) {
      this.toast.showError('Error', 'Please fill all required fields correctly');
      return;
    }

    this.testingConnection = true;
    const provider: UtmIdentityProvider = this.providerForm.value;

    this.providerService.testConnection(provider).subscribe({
      next: (res) => {
        if (res.body.success) {
          this.toast.showSuccess('Connection successful!');
        } else {
          this.toast.showError('Error', 'Connection test failed');
        }
        this.testingConnection = false;
      },
      error: (err) => {
        this.toast.showError('Error', 'An error occurred while testing the connection');
        this.testingConnection = false;
      }
    });
  }
}
