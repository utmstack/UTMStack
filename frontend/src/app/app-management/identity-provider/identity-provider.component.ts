import { Component, OnInit } from '@angular/core';
import { FormGroup } from '@angular/forms';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {
  ModalConfirmationComponent
} from '../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {
  IdentityProviderModalComponent
} from './shared/components/identity-provider-modal/identity-provider-modal.component';
import {UtmIdentityProvider} from './shared/models/utm-identity-provider.model';
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
    this.providerService.query({page: 0, size: 10}).subscribe({
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
    const deleteModalRef = this.modalService.open(ModalConfirmationComponent, {centered: true});

    deleteModalRef.componentInstance.header = 'Confirm delete operation';
    deleteModalRef.componentInstance.message = 'Are you sure that you want to delete the provider: ' + provider.providerType;
    deleteModalRef.componentInstance.confirmBtnText = 'Delete';
    deleteModalRef.componentInstance.confirmBtnIcon = 'icon-database-remove';
    deleteModalRef.componentInstance.confirmBtnType = 'delete';
    deleteModalRef.result.then(() => {

      this.providerService.delete(provider.id).subscribe({
        next: () => {
          this.toast.showSuccess('Provider deleted successfully');
          this.loadProviders();
        },
        error: () => {
          this.toast.showError( 'Error', 'An error occurred while deleting the provider');
        }
      });

    });
  }

  toggleActive(provider: UtmIdentityProvider): void {
    this.providerService.update(provider).subscribe({
      next: () => {
        this.toast.showSuccess(`Provider ${provider.active ? 'activated' : 'deactivated'}`);
        this.loadProviders();
      },
      error: () => {
        this.toast.showError( 'Error', 'An error occurred while updating provider status');
      }
    });
  }
}
