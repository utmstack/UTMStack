import {Component, EventEmitter, Input, Output} from '@angular/core';
import {PROVIDER_ICONS, ProviderType, UtmIdentityProvider} from '../../models/utm-identity-provider.model';

@Component({
  selector: 'app-provider-card',
  templateUrl: './provider.component.html',
  styleUrls: ['./provider.component.scss']
})
export class ProviderComponent {
  @Input() provider: UtmIdentityProvider;
  @Output() edit = new EventEmitter<UtmIdentityProvider>();
  @Output() delete = new EventEmitter<UtmIdentityProvider>();
  @Output() toggleStatus = new EventEmitter<UtmIdentityProvider>();

  getProviderIcon(type: ProviderType): string {
    return PROVIDER_ICONS[type] || 'icon-key';
  }

  openModal(provider: UtmIdentityProvider): void {
    this.edit.emit(provider);
  }

  deleteProvider(provider: UtmIdentityProvider): void {
    this.delete.emit(provider);
  }

  toggleActive(provider: UtmIdentityProvider): void {
    this.toggleStatus.emit({
      ...provider,
      active: !provider.active
    });
  }
}
