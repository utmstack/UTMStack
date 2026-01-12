import { Injectable } from '@angular/core';
import { SubscriptionLimitModalComponent } from '../components/subscription-limit-modal/subscription-limit-modal.component';
import { ModalService } from 'src/app/core/modal/modal.service';

export interface SubscriptionLimitModalOptions {
  title?: string;
  message: string;
  icon?: 'warning' | 'error' | 'info';
  showUpgradeButton?: boolean;
}

@Injectable({
  providedIn: 'root'
})
export class SubscriptionLimitModalService {

  constructor(private modalService: ModalService) { }

  showFeatureLimitModal(entity: string): void {
    const numericLimits = ['active projects', 'ai generations per month', 'allowed users', 'max resources'];
    const featureName = entity.replace(/_/g, ' ');

    if (numericLimits.includes(entity)) {
      let title = `${featureName.charAt(0).toUpperCase() + featureName.slice(1)} Limit Reached`;

      if (entity === 'allowed_users') {
        title = 'User Limit Reached';
      } else if (entity === 'active_projects') {
        title = 'Project Limit Reached';
      }

      this.showLimitModal({
        title: title,
        message: `You've reached your ${featureName} limit. Upgrade your plan to increase it.`,
        showUpgradeButton: true,
        icon: 'warning'
      });
    } else {
      this.showLimitModal({
        title: 'Feature Not Available',
        message: `The ${featureName} feature is not available in your current plan. Upgrade to access this feature.`,
        icon: 'info',
        showUpgradeButton: true
      });
    }
  }

  private showLimitModal(options: SubscriptionLimitModalOptions): void {
    const modalRef = this.modalService.open(SubscriptionLimitModalComponent, {
      centered: true
    });
    modalRef.componentInstance.title = options.title || 'Subscription Limit Reached';
    modalRef.componentInstance.message = options.message;
    modalRef.componentInstance.icon = options.icon || 'warning';
    modalRef.componentInstance.showUpgradeButton = options.showUpgradeButton || false;
  }
}
