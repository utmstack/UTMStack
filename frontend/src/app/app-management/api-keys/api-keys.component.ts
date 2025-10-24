import {Component, OnInit, TemplateRef, ViewChild} from '@angular/core';
import {NgbModal, NgbModalRef} from '@ng-bootstrap/ng-bootstrap';
import * as moment from 'moment';
import {UtmToastService} from "../../shared/alert/utm-toast.service";
import {
  ModalConfirmationComponent
} from "../../shared/components/utm/util/modal-confirmation/modal-confirmation.component";
import {ITEMS_PER_PAGE} from '../../shared/constants/pagination.constants';
import {SortEvent} from '../../shared/directives/sortable/type/sort-event';
import { ApiKeyModalComponent } from './shared/components/api-key-modal/api-key-modal.component';
import { ApiKeyResponse } from './shared/models/ApiKeyResponse';
import { ApiKeysService } from './shared/service/api-keys.service';

@Component({
  selector: 'app-api-keys',
  templateUrl: './api-keys.component.html',
  styleUrls: ['./api-keys.component.scss']
})
export class ApiKeysComponent implements OnInit {

  generating: string[] = [];
  apiKeys: ApiKeyResponse[] = [];
  loading = false;
  generatedApiKey = '';
  @ViewChild('generatedModal') generatedModal!: TemplateRef<any>;
  request =
    {
      sort: 'createdAt,desc',
      page: 0,
      size: ITEMS_PER_PAGE
    };
  generatedModalRef!: NgbModalRef;
  copied = false;

  constructor( private toastService: UtmToastService,
               private apiKeyService: ApiKeysService,
               private modalService: NgbModal
  ) {}

  ngOnInit(): void {
    this.loadKeys();
  }

  loadKeys(): void {
    this.loading = true;
    this.apiKeyService.list().subscribe({
      next: (res) => {
        this.apiKeys = res.body || [];
        this.loading = false;
      },
      error: () =>  {
        this.loading = false;
        this.apiKeys = [];
      }
    });
  }

  copyToClipboard(): void {
    (navigator as any).clipboard.writeText(this.generatedApiKey).then(() => {
      this.copied = true;
    });
  }

  openCreateModal(): void {
    const modalRef = this.modalService.open(ApiKeyModalComponent, { centered: true });

    modalRef.result.then((key: ApiKeyResponse) => {
      if (key) {
        this.generateKey(key);
      }
    });
  }

  editKey(key: ApiKeyResponse): void {
    const modalRef = this.modalService.open(ApiKeyModalComponent, {centered: true});
    modalRef.componentInstance.apiKey = key;

    modalRef.result.then((key: ApiKeyResponse) => {
      if (key) {
        this.generateKey(key);
      }
    });
  }

  deleteKey(apiKey: ApiKeyResponse): void {
    const modalRef = this.modalService.open(ModalConfirmationComponent, {centered: true});
    modalRef.componentInstance.header = `Delete API Key: ${apiKey.name}`;
    modalRef.componentInstance.message = 'Are you sure you want to delete this API key?';
    modalRef.componentInstance.confirmBtnType = 'delete';
    modalRef.componentInstance.type = 'danger';
    modalRef.componentInstance.confirmBtnText = 'Delete';
    modalRef.componentInstance.confirmBtnIcon = 'icon-cross-circle';

    modalRef.result.then(reason => {
      if (reason === 'ok') {
        this.delete(apiKey);
      }
    });
  }

  delete(apiKey: ApiKeyResponse): void {
    this.apiKeyService.delete(apiKey.id).subscribe({
      next: () => {
        this.toastService.showSuccess('API key deleted successfully.');
        this.loadKeys();
      },
      error: (err) => {
        this.toastService.showError('Error', 'An error occurred while deleting the API key.');
        throw err;
      }
    });
  }

  getDaysUntilExpire(expiresAt: string): number {
    if (!expiresAt) {
      return 0;
    }
    const today = moment();
    const expireDate = moment(expiresAt);
    return expireDate.diff(today, 'days');
  }

  onSortBy($event: SortEvent | string) {

  }

  maskSecrets(str: string): string {
    if (!str || str.length <= 10) {
      return str;
    }
    const prefix = str.substring(0, 10);
    const maskLength = str.length - 30;
    const maskedPart = '*'.repeat(maskLength);
    return prefix + maskedPart;
  }

  generateKey(apiKey: ApiKeyResponse): void {
    this.generating.push(apiKey.id);
    this.apiKeyService.generateApiKey(apiKey.id).subscribe(response => {
      this.generatedApiKey = response.body ? response.body : "";
      this.generatedModalRef = this.modalService.open(this.generatedModal, {centered: true});
      const index = this.generating.indexOf(apiKey.id);
      if (index > -1) {
        this.generating.splice(index, 1);
      }
      this.loadKeys();
    });
  }

  isApiKeyExpired(expiresAt?: string | null ): boolean {
    if (!expiresAt) {
      return false;
    }
    const expirationTime = new Date(expiresAt).getTime();
    return expirationTime < Date.now();
  }

  close() {
    this.generatedModalRef.close();
    this.copied = false;
    this.generatedApiKey = '';
  }
}
