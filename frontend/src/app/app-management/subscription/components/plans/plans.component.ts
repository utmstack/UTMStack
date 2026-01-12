import {
  Component,
  Input,
  OnInit,
  ChangeDetectionStrategy,
  ChangeDetectorRef,
  OnDestroy,
} from '@angular/core';
import {
  combineLatest,
  Observable,
  Subject,
} from 'rxjs';
import {
  takeUntil,
  map,
  finalize,
} from 'rxjs/operators'
import { Router } from '@angular/router';
import {
  PlanModel,
  PlanPrice,
  StripeUrlModel,
  SubscriptionModel,
} from '../../models/plan.model';
import { SubscriptionService } from '../../services/subscription.service';

type Status = 'new' | 'trialing' | 'active' | 'from_code';

@Component({
  selector: 'app-plan-cards',
  templateUrl: './plans.component.html',
  styleUrls: ['./plans.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class PlanCardsComponent implements OnInit, OnDestroy {
  @Input() mode: 'profile' | 'setup' = 'profile';

  plans$!: Observable<PlanModel[]>;
  subscription$!: Observable<SubscriptionModel | undefined>;

  codeSubscriptionLoading: boolean = false;
  codeSubscriptionError: string = '';

  subStatus: Status = 'new';
  currentPriceId: string | null = null;
  currentPlanPosition: number | null = null;

  readonly customPlanUrl = 'https://example.com';
  private destroy$ = new Subject<void>();

  constructor(
    private subscriptionService: SubscriptionService,
    private cdr: ChangeDetectorRef,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.plans$ = this.subscriptionService.plansObservable.pipe(
      map((plans) => plans.filter((p) => p.position >= 0))
    );
    this.subscription$ = this.subscriptionService.subscriptionObservable;

    combineLatest([this.plans$, this.subscription$])
      .pipe(takeUntil(this.destroy$))
      .subscribe(([plans, sub]:[PlanModel[],SubscriptionModel]) => {
        if (!sub || sub.status === 'new') {
          this.subStatus = 'new';
          this.currentPriceId = null;
          this.currentPlanPosition = null;
        } else {
          this.subStatus = sub.status as Status;
          this.currentPriceId = sub.price_id;
        }

        if (this.currentPriceId != null) {
          const currentPlan = plans.filter((plan) =>
            plan.prices.find((p) => p.id == this.currentPriceId)
          )[0];
          if (currentPlan) {
            this.currentPlanPosition = currentPlan.position;
            const currentPrice = currentPlan.prices.find(
              (p) => p.id == this.currentPriceId
            );
            if (currentPrice.interval && currentPrice.interval == 'year') {
              this.isAnnualSelected[currentPlan.id] = true;
            }
          }
        }

        this.cdr.markForCheck();
      });
  }

  isAnnualSelected: { [planId: string]: boolean } = {};

  setBillingPeriod(planId: string, isAnnual: boolean) {
    this.isAnnualSelected[planId] = isAnnual;
  }

  getButtonText(plan: PlanModel): string {
    if (this.subStatus === 'new' || this.subStatus === 'from_code') {
      return 'Select Plan';
    }

    const selectedPrice = this.getSelectedPrice(plan);
    if (selectedPrice && selectedPrice.id === this.currentPriceId) {
      return 'Manage Plan';
    }

    if (this.currentPlanPosition !== null) {
      return plan.position > this.currentPlanPosition ? 'Upgrade' : 'Downgrade';
    }

    return 'Select Plan';
  }

  getSelectedPrice(plan: PlanModel): PlanPrice | undefined {
    // If plan has only one price (like Free plan with one_time), return that price
    if (plan.prices.length === 1) {
      return plan.prices[0];
    }

    // Free plans always show yearly price
    if (plan.name.toLowerCase().includes('free')) {
      return plan.prices.find((p) => p.interval === 'year') || plan.prices[0];
    }

    // For plans with multiple prices, use the toggle selection logic
    return plan.prices.find((p) =>
      this.isAnnualSelected[plan.id]
        ? p.interval === 'year'
        : p.interval === 'month'
    );
  }

  onPlanAction(plan: PlanModel): void {
    const selectedPrice = this.getSelectedPrice(plan);
    if (!selectedPrice) {
      console.error('No selected price found for plan:', plan);
      return;
    }

    // Si es el plan actual, ir al portal para gestionar
    if (selectedPrice.id === this.currentPriceId) {
      this.subscriptionService
        .getPortalSession()
        .subscribe((r: StripeUrlModel) => (window.location.href = r.url));
    } else {
      // Para cualquier otro plan (nuevo usuario o upgrade/downgrade), usar checkout
      this.subscriptionService
        .getCheckoutSession(selectedPrice.id)
        .subscribe((r: StripeUrlModel) => (window.location.href = r.url));
    }
  }

  contactSupport(): void {
    window.open(this.customPlanUrl, '_blank');
  }

  isPromoCodeModalOpen = false;
  promoCode = '';

  openPromoCodeModal() {
    this.codeSubscriptionError = '';
    this.isPromoCodeModalOpen = true;
  }

  closePromoCodeModal() {
    this.isPromoCodeModalOpen = false;
    this.promoCode = '';
    this.codeSubscriptionError = '';
  }

  applyPromoCode() {
    if (this.promoCode) {
      this.codeSubscriptionLoading = true;
      this.subscriptionService
        .subscribeWithPromoCode(this.promoCode)
        .pipe(finalize(() => (this.codeSubscriptionLoading = false)))
        .subscribe({
          next: () => {
            this.closePromoCodeModal();
            if (this.subStatus == 'new') {
              this.router.navigate(['/onboarding']);
            } else {
              this.router.navigate(['/']);
            }
          },
          error: (error) => {
            console.error('Failed to apply promo code', error);
            this.codeSubscriptionError = 'Invalid Code';
            this.codeSubscriptionLoading = false;
            this.cdr.markForCheck();
          },
        });
    }
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}

