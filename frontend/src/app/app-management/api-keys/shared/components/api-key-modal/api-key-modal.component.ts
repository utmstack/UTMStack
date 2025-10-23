import {Component, Input, OnInit} from '@angular/core';
import { NgbActiveModal } from '@ng-bootstrap/ng-bootstrap';
import { ApiKeysService } from '../../service/api-keys.service';
import { FormBuilder, FormGroup, Validators, FormArray } from '@angular/forms';
import {ApiKeyUpsert} from '../../models/ApiKeyUpsert';
import {IpFormsValidators} from "../../../../../rule-management/app-rule/validators/ip.forms.validators";

@Component({
  selector: 'app-api-key-modal',
  templateUrl: './api-key-modal.component.html',
  styleUrls: ['./api-key-modal.component.scss']
})
export class ApiKeyModalComponent implements OnInit {

  @Input() apiKey: ApiKeyUpsert = null;

  apiKeyForm: FormGroup;
  ipInput = '';
  loading = false;
  errorMsg = '';
  isSaving: string | string[] | Set<string> | { [p: string]: any };
  ipInputError: string = '';

  constructor(
    public activeModal: NgbActiveModal,
    private apiKeyService: ApiKeysService,
    private fb: FormBuilder) {

    this.apiKeyForm = this.fb.group({
      name: ['', Validators.required],
      allowedIp: this.fb.array([]),
      expiresAt: ['', Validators.required],
    });

  }

  ngOnInit(): void {}

  get allowedIp(): FormArray {
    return this.apiKeyForm.get('allowedIp') as FormArray;
  }

  addIp(): void {
    const trimmedIp = this.ipInput.trim();

    if (!trimmedIp) {
      this.ipInputError = 'Please enter an IP address or CIDR'; // Se asigna el error
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
    const payload = {
      ...this.apiKeyForm.value,
      expiresAt: this.apiKeyForm.value.expiresAt + ':00.000Z',
    };
    this.apiKeyService.create(payload).subscribe({
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

  getIpType(value: string): string {
    if (!value) { return ''; }
    if (value.includes('/')) {
      return value.includes(':') ? 'IPv6 CIDR' : 'IPv4 CIDR';
    }
    return value.includes(':') ? 'IPv6' : 'IPv4';
  }
}

