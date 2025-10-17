import { Component, OnInit } from '@angular/core';
import { NgbActiveModal } from '@ng-bootstrap/ng-bootstrap';
import { ApiKeysService } from '../../service/api-keys.service';
import { FormBuilder, FormGroup, Validators, FormArray } from '@angular/forms';

@Component({
  selector: 'app-api-key-modal',
  templateUrl: './api-key-modal.component.html',
  styleUrls: ['./api-key-modal.component.scss']
})
export class ApiKeyModalComponent implements OnInit {
  apiKeyForm: FormGroup;
  ipInput = '';
  loading = false;
  errorMsg = '';

  constructor(
    public activeModal: NgbActiveModal,
    private apiKeyService: ApiKeysService,
    private fb: FormBuilder
  ) {
    this.apiKeyForm = this.fb.group({
      name: ['', Validators.required],
      allowedIp: this.fb.array([]),
      expiresAt: [null]
    });
  }

  ngOnInit(): void {}

  get allowedIp(): FormArray {
    return this.apiKeyForm.get('allowedIp') as FormArray;
  }

  addIp(): void {
    const ip = this.ipInput.trim();
    if (ip) {
      this.allowedIp.push(this.fb.control(ip));
      this.ipInput = '';
    }
  }

  removeIp(index: number): void {
    this.allowedIp.removeAt(index);
  }

  create(): void {
    this.errorMsg = '';

    if (this.apiKeyForm.invalid) {
      this.errorMsg = 'Name is required.';
      return;
    }

    this.loading = true;
    this.apiKeyService.create(this.apiKeyForm.value).subscribe({
      next: () => {
        this.loading = false;
        this.activeModal.close('created');
      },
      error: (err) => {
        this.loading = false;
        this.errorMsg = err.error.message || 'An error occurred while creating the API key.';
      }
    });
  }
}

