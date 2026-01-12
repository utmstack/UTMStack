import {
  Directive,
  Input,
  OnDestroy,
  Renderer2,
  HostListener,
  ElementRef,
  Output,
  EventEmitter
} from '@angular/core';
import { Subject, of } from 'rxjs';
import { switchMap, takeUntil } from 'rxjs/operators';
import { SubscriptionService } from '../services/subscription.service';
import { SubscriptionLimitModalService } from '../services/subscription-limit-modal.service';

interface CheckConfig {
  entity: string;
  projectId?: string;
  versionId?: string;
}

@Directive({
  selector: '[allowedInSubscription]',
})
export class SubscriptionAllowedDirective implements OnDestroy {
  private config?: CheckConfig;
  private isFeatureAllowed: boolean = true;
  private destroy$ = new Subject<void>();

  @Output() allowedClick = new EventEmitter<MouseEvent>();

  constructor(
    private elementRef: ElementRef,
    private renderer: Renderer2,
    private subscriptionService: SubscriptionService,
    private modalService: SubscriptionLimitModalService
  ) {}

  @Input()
  set allowedInSubscription(config: CheckConfig) {
    if (!config || !config.entity) return;

    // Check if config actually changed to avoid unnecessary re-checks
    const configChanged = !this.config ||
                         this.config.entity !== config.entity ||
                         this.config.projectId !== config.projectId ||
                         this.config.versionId !== config.versionId;

    this.config = config;

    // Only check if config changed
    if (configChanged) {
      this.checkFeatureAccess();
    }
  }

  @HostListener('click', ['$event'])
  onClick(event: MouseEvent): void {
    if (this.isFeatureAllowed) {
      this.allowedClick.emit(event);
    } else {
      event.preventDefault();
      event.stopPropagation();
      event.stopImmediatePropagation();
      this.modalService.showFeatureLimitModal(this.config!.entity.replace(/_/g, ' '));
    }
  }

  private checkFeatureAccess(): void {
    this.subscriptionService.subscriptionObservable
      .pipe(
        switchMap(() => {
          const { entity, projectId, versionId } = this.config!;

          const featuresNeedingProjectId = ['ai_requests', 'max_resources'];

          if (featuresNeedingProjectId.includes(entity)) {
            // If projectId is required but not available, skip the validation
            if (!projectId) {
              console.warn(`Feature ${entity} requires projectId but it's not provided`);
              return of(true); // Allow by default until projectId is available
            }
            return this.subscriptionService.validateFeature(entity, projectId, versionId);
          }

          return this.subscriptionService.validateFeature(entity);
        }),
        takeUntil(this.destroy$)
      )
      .subscribe(isAllowed => {
        this.isFeatureAllowed = isAllowed;
        this.updateElementStyle(isAllowed);
      });
  }

  private updateElementStyle(isAllowed: boolean): void {
    const element = this.elementRef.nativeElement;
    if (!element) return;

    if (!isAllowed) {
      this.renderer.setStyle(element, 'opacity', '0.5');
      this.renderer.setStyle(element, 'cursor', 'not-allowed');
    } else {
      this.renderer.removeStyle(element, 'opacity');
      this.renderer.removeStyle(element, 'cursor');
    }
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
