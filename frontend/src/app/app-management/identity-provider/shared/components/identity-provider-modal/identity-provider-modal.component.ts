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

    const formData = this.convertToFormData(formValue, this.providerFormComponent.privateKeyFile,
      this.providerFormComponent.certificateFile);

    const request = this.editMode && this.provider.id
      ? this.providerService.update(formData)
      : this.providerService.create(formData);

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

  onChangePrivateCertificateFile($event: { file: File; name: string }): void {
    this.providerFormComponent.certificateFile = $event.file;
    this.providerFormComponent.certificateFileName = $event.name;
  }

  private convertToFormData(data: any, privateKey?: File | null, certificate?: File | null): FormData {
    const formData = new FormData();

    // Agregar campos del formulario
    formData.append('name', data.name);
    formData.append('providerType', data.providerType);
    formData.append('metadataUrl', data.metadataUrl);
    formData.append('active', data.active.toString());

    // Agregar archivos si existen
    if (privateKey) {
      formData.append('spPrivateKeyFile', privateKey);
    }
    if (certificate) {
      formData.append('spCertificateFile', certificate);
    }

    return formData;
  }
}
