import {HttpClient} from "@angular/common/http";
import { Component, EventEmitter, Input, OnInit, Output } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { ProviderType, UtmIdentityProvider } from '../../models/utm-identity-provider.model';
import {validateMetadataUrl} from '../../validators/validator';

@Component({
  selector: 'app-provider-form',
  templateUrl: './provider-form.component.html',
  styleUrls: ['./provider-form.component.scss']
})
export class ProviderFormComponent implements OnInit {
  @Input() provider?: UtmIdentityProvider;
  @Input() loading = false;
  @Input() testingConnection = false;
  @Input() providers: UtmIdentityProvider[] = [];

  @Output() save = new EventEmitter<FormData>();
  @Output() test = new EventEmitter<void>();
  @Output() cancel = new EventEmitter<void>();

  editMode = false;
  providerForm!: FormGroup;
  privateKeyFile?: File;
  certificateFile?: File;
  privateKeyFileName = '';
  certificateFileName = '';

  providerTypes: { label: string; value: ProviderType }[] = [];

  spEntityId = '';
  spAcsUrl = '';

  constructor(private fb: FormBuilder,
              private http: HttpClient) {}

  ngOnInit(): void {
    this.editMode = !!this.provider;
    this.providerTypes = Object.values(ProviderType)
      .filter(type => !this.providers.some(p => p.providerType === type))
      .map((value) => ({
        label: value.charAt(0) + value.slice(1).toLowerCase(),
        value
      }));
    this.initForm();

    if (this.editMode && this.provider) {
      const { spPrivateKeyPem, ...providerData } = this.provider;
      this.providerForm.patchValue(providerData);
      this.makePrivateKeyOptional();
    }
    this.generateSpIdentifiers();
  }

  initForm(): void {
    this.providerForm = this.fb.group({
      name: ['', [Validators.required, Validators.minLength(3)]],
      providerType: [ProviderType.GOOGLE, Validators.required],
      metadataUrl: ['', [Validators.required, Validators.pattern(/^https?:\/\/.+/)], [validateMetadataUrl(this.http)]],
      active: [true]
    });
  }

  private makePrivateKeyOptional(): void {
    this.providerForm.setControl('spPrivateKeyPem', this.fb.control(''));
  }

  private generateSpIdentifiers(): void {
    const origin = window.location.origin;
    const provider = this.providerForm.get('providerType') ? this.providerForm.get('providerType').value : ProviderType.GOOGLE;
    this.spEntityId = `${origin}/saml/sp`;
    this.spAcsUrl = `${origin}/login/saml2/sso/${provider.toLowerCase()}`;
  }

  onPrivateKeySelected(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
      this.privateKeyFile = input.files[0];
      this.privateKeyFileName = this.privateKeyFile.name;
    }
  }

  onCertificateSelected(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
      this.certificateFile = input.files[0];
      this.certificateFileName = this.certificateFile.name;
    }
  }

  clearPrivateKeyFile(): void {
    this.privateKeyFile = undefined;
    this.privateKeyFileName = '';
  }

  clearCertificateFile(): void {
    this.certificateFile = undefined;
    this.certificateFileName = '';
  }

  /*saveProvider(): void {
    if (!this.providerForm.valid) {
      return;
    }

    // En creación, ambos archivos son requeridos
    if (!this.editMode && (!this.privateKeyFile || !this.certificateFile)) {
      console.error('Both private key and certificate files are required');
      return;
    }

    // En edición, al menos uno debe estar presente si se van a actualizar
    if (this.editMode && !this.privateKeyFile && !this.certificateFile) {
      // Si no hay archivos nuevos, solo enviar datos del formulario
      const formValue: UtmIdentityProvider = this.providerForm.value;
      this.save.emit(this.convertToFormData(formValue, null, null));
      return;
    }


    this.save.emit(formData);
  }*/

  testConnection(): void {
    this.test.emit();
  }

  cancelForm(): void {
    this.cancel.emit();
  }

  copyToClipboard(label: string, value: string): void {
    if (!value) {
      return;
    }

    (navigator as any).clipboard.writeText(value).then(
      () => {
        console.log(`${label} copied to clipboard`);
      },
      (err) => {
        console.error('Error copying to clipboard:', err);
      }
    );
  }
}
