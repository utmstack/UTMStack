import {Component, Input, OnInit, ViewChild} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmIdentityProvider} from '../../models/utm-identity-provider.model';
import {UtmIdentityProviderService} from '../../services/utm-identity-provider.service';
import {ProviderFormComponent} from '../provider-form/provider-form.component';

@Component({
  selector: 'app-identity-provider-modal',
  templateUrl: './identity-provider-modal.component.html',
  styleUrls: ['./identity-provider-modal.component.scss']
})
export class IdentityProviderModalComponent implements OnInit {
  @Input() provider?: UtmIdentityProvider;
  @Input() editMode = false;

  @ViewChild('providerForm') providerFormComponent!: ProviderFormComponent;

  selectedProvider?: UtmIdentityProvider;
  loading = false;
  testingConnection = false;

  constructor(
    public activeModal: NgbActiveModal,
    private providerService: UtmIdentityProviderService
  ) {}

  ngOnInit(): void {
    this.selectedProvider = this.provider;
  }

  saveProvider(): void {

    if (!this.providerFormComponent.providerForm.valid) {
      this.markFormGroupTouched(this.providerFormComponent.providerForm);
      return;
    }

    this.loading = true;

    const formValue = this.providerFormComponent.providerForm.value;
    const providerData: UtmIdentityProvider = {
      ...formValue,
      id: this.provider.id,
      scopes: Array.isArray(formValue.scopes) ? formValue.scopes.join(',') : formValue.scopes,
      allowedDomains: Array.isArray(formValue.allowedDomains)
        ? formValue.allowedDomains.join(',')
        : formValue.allowedDomains
    };

    const request = this.editMode && this.provider.id
      ? this.providerService.update(providerData)
      : this.providerService.create(providerData);

    request.subscribe({
      next: () => {
        this.loading = false;
        this.activeModal.close(true);
      },
      error: (error) => {
        this.loading = false;
        console.error('Error saving provider:', error);
      }
    });
  }

  testConnection(): void {
    if (this.providerFormComponent.providerForm && !this.providerFormComponent.providerForm.valid) {
      // ✅ Usar función helper para marcar todos los controles como touched
      this.markFormGroupTouched(this.providerFormComponent.providerForm);
      return;
    }

    this.testingConnection = true;

    this.providerService.testConnection(this.providerFormComponent.providerForm.value).subscribe({
      next: () => {
        this.testingConnection = false;
        alert('✅ Connection test successful!');
      },
      error: (error) => {
        this.testingConnection = false;
        alert('❌ Connection test failed: ' + error.message);
      }
    });
  }

  closeModal(): void {
    this.activeModal.dismiss();
  }

  private markFormGroupTouched(formGroup: any): void {
    Object.keys(formGroup.controls).forEach(key => {
      const control = formGroup.get(key);
      control.markAsTouched();

      if (control.controls) {
        this.markFormGroupTouched(control);
      }
    });
  }
}
