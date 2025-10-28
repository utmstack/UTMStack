import {HttpErrorResponse} from '@angular/common/http';
import {Component, Input, OnInit} from '@angular/core';
import { FormArray, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { NgbActiveModal } from '@ng-bootstrap/ng-bootstrap';
import {IpFormsValidators} from '../../../../../rule-management/app-rule/validators/ip.forms.validators';
import {UtmToastService} from '../../../../../shared/alert/utm-toast.service';
import {ApiKeyResponse} from '../../models/ApiKeyResponse';
import { ApiKeysService } from '../../service/api-keys.service';

@Component({
  selector: 'app-api-key-modal',
  templateUrl: './api-key-modal.component.html',
  styleUrls: ['./api-key-modal.component.scss']
})
export class ApiKeyModalComponent implements OnInit {

  @Input() apiKey: ApiKeyResponse = null;

  apiKeyForm: FormGroup;
  ipInput = '';
  loading = false;
  errorMsg = '';
  isSaving: string | string[] | Set<string> | { [p: string]: any };
  minDate = { year: new Date().getFullYear(), month: new Date().getMonth() + 1, day: new Date().getDate() };
  ipInputError = '';

  constructor(  public activeModal: NgbActiveModal,
                private apiKeyService: ApiKeysService,
                private fb: FormBuilder,
                private toastService: UtmToastService) {
  }

  ngOnInit(): void {

    const expiresAtDate = this.apiKey && this.apiKey.expiresAt ? new Date(this.apiKey.expiresAt) : null;
    const expiresAtNgbDate = expiresAtDate ? {
      year: expiresAtDate.getUTCFullYear(),
      month: expiresAtDate.getUTCMonth() + 1,
      day: expiresAtDate.getUTCDate()
    } : null;

    this.apiKeyForm = this.fb.group({
      name: [ this.apiKey ? this.apiKey.name : '', Validators.required],
      allowedIp: this.fb.array(this.apiKey ? this.apiKey.allowedIp : []),
      expiresAt: [expiresAtNgbDate, Validators.required],
    });
  }

  get allowedIp(): FormArray {
    return this.apiKeyForm.get('allowedIp') as FormArray;
  }

  addIp(): void {
    const trimmedIp = this.ipInput.trim();

    if (!trimmedIp) {
      this.ipInputError = 'Please enter an IP address or CIDR'; // Error is assigned
      return;
    }

    const tempControl = this.fb.control(trimmedIp, [IpFormsValidators.ipOrCidr()]);

    if (tempControl.invalid) {
      if (tempControl.hasError('invalidIp')) {
        this.ipInputError = 'Invalid IP address format';
      } else if (tempControl.hasError('invalidCidr')) {
        this.ipInputError = 'Invalid CIDR format';
      }
      return;
    }

    const isDuplicate = this.allowedIp.controls.some(
      control => control.value === trimmedIp
    );

    if (isDuplicate) {
      this.ipInputError = 'This IP is already added';
      return;
    }

    this.allowedIp.push(this.fb.control(trimmedIp, [IpFormsValidators.ipOrCidr()]));
    this.ipInput = '';
    this.ipInputError = '';
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

    const rawDate = this.apiKeyForm.get('expiresAt').value;
    let formattedDate = rawDate;

    if (rawDate && typeof rawDate === 'object') {
      formattedDate = `${rawDate.year}-${String(rawDate.month).padStart(2, '0')}-${String(rawDate.day).padStart(2, '0')}T00:00:00.000Z`;
    }

    const payload = {
      ...this.apiKeyForm.value,
      expiresAt: formattedDate,
    };

    const save = this.apiKey ? this.apiKeyService.update(this.apiKey.id, payload) :
      this.apiKeyService.create(payload);

    save.subscribe({
      next: (response) => {
        this.loading = false;
        this.activeModal.close(response.body as ApiKeyResponse);
      },
      error: (err: HttpErrorResponse) => {
        this.loading = false;
        if (err.status === 409) {
          this.toastService.showError('Error', 'An API key with this name already exists.');
        } else if (err.status === 500) {
          this.toastService.showError('Error', 'Server error occurred while creating the API key.');
        }
      }
    });
  }

  getIpType(value: string): string {
    if (!value) { return ''; }
    if (value.includes('/')) {
      return value.includes(':') ? 'IPv6 CIDR' : 'IPv4 CIDR';
    }
    return value.includes(':') ? 'IPv6' : 'IPv4';
  }
}

