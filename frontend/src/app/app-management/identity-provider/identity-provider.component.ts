import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {ProviderFormComponent} from './shared/components/provider-form/provider-form.component';
import {UtmIdentityProvider} from './shared/models/utm-identity-provider.model';
import {UtmIdentityProviderService} from './shared/services/utm-identity-provider.service';
import {
  IdentityProviderModalComponent
} from "./shared/components/identity-provider-modal/identity-provider-modal.component";

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
  selectedProvider: UtmIdentityProvider | null = null;

  constructor(
    private modalService: NgbModal,
    private providerService: UtmIdentityProviderService,
    private toast: UtmToastService) {
  }

  ngOnInit(): void {
    this.loadProviders();
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
    const modalRef = this.modalService.open(IdentityProviderModalComponent, {
      size: 'lg',
      centered: true,
    });

    modalRef.componentInstance.provider = provider;
    modalRef.componentInstance.editMode = !!provider;

    modalRef.result.then(
      (result) => {
        if (result) {
          this.loadProviders();
        }
      },
      () => {
        // Modal dismissed
      }
    );
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
}
