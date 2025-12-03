import {HttpErrorResponse} from "@angular/common/http";
import {Component, Input, OnInit, ViewChild} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../../shared/alert/utm-toast.service';
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
  @Input() providers: UtmIdentityProvider[] = [];

  @ViewChild('providerForm') providerFormComponent!: ProviderFormComponent;

  selectedProvider?: UtmIdentityProvider;
  loading = false;
  testingConnection = false;

  constructor(public activeModal: NgbActiveModal,
              private providerService: UtmIdentityProviderService,
              private toastService: UtmToastService
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

    const request = this.editMode && this.provider.id
      ? this.providerService.update(this.provider.id, formValue,
        this.providerFormComponent.privateKeyFile, this.providerFormComponent.certificateFile)
      : this.providerService.create(formValue, this.providerFormComponent.privateKeyFile, this.providerFormComponent.certificateFile);

    request.subscribe({
      next: () => {
        this.loading = false;
        this.activeModal.close(true);
        this.toastService.showSuccessProcess('Success', `Provider ${this.editMode ? 'updated' : 'created'} successfully.`);
      },
      error: (error: HttpErrorResponse) => {
        this.loading = false;

        if (error.status === 400) {
          this.toastService.showError('Validation Error', 'Please check the form for errors and try again.');
        } else {
          this.toastService.showError('Error',
            `An error occurred while ${this.editMode ? 'updating' : 'creating'} the provider`);
        }
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
        this.toastService.showSuccessProcess('Success', 'Connection test succeeded.');
      },
      error: (error) => {
        this.testingConnection = false;
        this.toastService.showError('Error', `Connection test failed`);
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
}
