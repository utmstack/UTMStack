import { Component, OnInit } from '@angular/core';
import { ApiKeysService } from './shared/service/api-keys.service';
import { ApiKeyResponse } from './shared/models/ApiKeyResponse';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import { ApiKeyModalComponent } from './shared/components/api-key-modal/api-key-modal.component';

@Component({
  selector: 'app-api-keys',
  templateUrl: './api-keys.component.html',
  styleUrls: ['./api-keys.component.scss']
})
export class ApiKeysComponent implements OnInit {
  apiKeys: ApiKeyResponse[] = [];
  loading = false;

  constructor(
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
      error: () => (this.loading = false)
    });
  }

  openCreateModal(): void {
    const modalRef = this.modalService.open(ApiKeyModalComponent, { centered: true, size: 'lg' });
    modalRef.result.then((result) => {
      if (result === 'created') this.loadKeys();
    }).catch(() => {});
  }

  deleteKey(id: string): void {
    if (!confirm('Are you sure you want to delete this API key?')) return;
    this.apiKeyService.delete(id).subscribe(() => this.loadKeys());
  }

  regenerateKey(id: string): void {
    this.apiKeyService.generate(id).subscribe((res) => {
      alert('New API Key: ' + res.body);
    });
  }
}
