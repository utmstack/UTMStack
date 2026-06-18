import {Component, EventEmitter, Input, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {FederationInstance} from '../../domain/federation-instance.model';
import {FederationInstanceInput, FederationInstancesService} from '../../services/federation-instances.service';

interface InstanceFormValue {
  name: string;
  baseUrl: string;
  apiKey: string;
  tlsSkipVerify: boolean;
}

@Component({
  selector: 'app-federation-instance-form-modal',
  templateUrl: './instance-form-modal.component.html',
  styleUrls: ['./instance-form-modal.component.scss']
})
export class InstanceFormModalComponent {
  @Input() instance: FederationInstance | null = null;
  @Output() saved = new EventEmitter<FederationInstance>();
  submitting = false;
  errorMessage: string | null = null;
  model: InstanceFormValue = {
    name: '',
    baseUrl: '',
    apiKey: '',
    tlsSkipVerify: false
  };

  constructor(public activeModal: NgbActiveModal,
              private instancesService: FederationInstancesService) {}

  ngOnInit(): void {
    if (this.instance) {
      this.model = {
        name: this.instance.name,
        baseUrl: this.instance.baseUrl,
        apiKey: '',
        tlsSkipVerify: this.instance.tlsSkipVerify
      };
    }
  }

  get isEditMode(): boolean {
    return this.instance !== null;
  }

  get title(): string {
    return this.isEditMode ? 'Edit instance' : 'Connect instance';
  }

  get apiKeyHint(): string {
    return this.isEditMode
      ? 'Leave empty to keep the existing API key.'
      : '';
  }

  submit(): void {
    if (this.submitting) {
      return;
    }
    this.submitting = true;
    this.errorMessage = null;
    const payload: FederationInstanceInput = {
      name: this.model.name,
      baseUrl: this.model.baseUrl,
      tlsSkipVerify: this.model.tlsSkipVerify
    };
    if (this.model.apiKey) {
      payload.apiKey = this.model.apiKey;
    }
    const action$ = this.isEditMode && this.instance
      ? this.instancesService.update(this.instance.id, payload)
      : this.instancesService.create(payload);
    action$.subscribe({
      next: instance => {
        this.submitting = false;
        this.saved.emit(instance);
      },
      error: err => {
        this.submitting = false;
        this.errorMessage = (err && err.error && err.error.message)
          || 'Failed to save instance.';
      }
    });
  }

  cancel(): void {
    this.activeModal.dismiss();
  }
}
