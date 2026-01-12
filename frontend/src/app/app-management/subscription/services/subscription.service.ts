import { Injectable, OnDestroy } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, Observable, Subscription, throwError, of } from 'rxjs';
import { catchError, finalize, map, tap } from 'rxjs/operators';

import { PlanModel, SubscriptionModel, StripeUrlModel } from '../models/plan.model';
import { FeatureValidation } from '../models/feature.validation';
import { environment } from 'src/environments/environment';

const API_SUBSCRIPTION_URL = `${environment.CUSTOMMER_MANAGER_API_URL}/subscription`;

@Injectable({
  providedIn: 'root',
})
export class SubscriptionService implements OnDestroy {
  private plans$ = new BehaviorSubject<PlanModel[]>([]);
  private subscription$ = new BehaviorSubject<SubscriptionModel | undefined>(undefined);
  private isLoading$ = new BehaviorSubject<boolean>(false);
  private subs: Subscription[] = [];

  constructor(private http: HttpClient) {
    this.fetchPlans().subscribe();
    this.fetchSubscription().subscribe();
  }

  // Public Observables
  get plansObservable(): Observable<PlanModel[]> {
    return this.plans$.asObservable();
  }

  get subscriptionObservable(): Observable<SubscriptionModel | undefined> {
    return this.subscription$.asObservable();
  }

  get loading$(): Observable<boolean> {
    return this.isLoading$.asObservable();
  }

  // Data Accessors
  get currentPlans(): PlanModel[] {
    return this.plans$.value;
  }

  get currentSubscription(): SubscriptionModel | undefined {
    return this.subscription$.value;
  }

  // API Calls
  fetchPlans(): Observable<PlanModel[]> {
    this.isLoading$.next(true);
    return this.http.get<PlanModel[]>(`${API_SUBSCRIPTION_URL}/plans`).pipe(
      tap((plans)=>{console.log(plans)}),
      tap((plans)=>{alert()}),
      tap((plans) => this.plans$.next(plans.sort((a, b) => a.position - b.position))),
      finalize(() => this.isLoading$.next(false)),
      catchError((err) => {
        console.error('Failed to fetch plans', err);
        return throwError(() => new Error('Failed to fetch plans'));
      })
    );
  }

  fetchSubscription(): Observable<SubscriptionModel> {
    this.isLoading$.next(true);
    return this.http.get<SubscriptionModel>(API_SUBSCRIPTION_URL).pipe(
      tap((sub) => {
        this.subscription$.next(sub);
      }),
      finalize(() => this.isLoading$.next(false)),
      catchError((err) => {
        console.error('Failed to fetch subscription', err);
        if (err.status === 404) {
          this.subscription$.next({ status: 'new' } as SubscriptionModel);
          return of({ status: 'new' } as SubscriptionModel);
        }
        return throwError(() => new Error('Failed to fetch subscription'));
      })
    );
  }

  getCheckoutSession(priceId: string): Observable<StripeUrlModel> {
    return this.http.get<StripeUrlModel>(`${API_SUBSCRIPTION_URL}/checkout?price_id=${priceId}`).pipe(
      catchError((err) => {
        console.error('Failed to create checkout session', err);
        return throwError(() => new Error('Checkout session failed'));
      })
    );
  }

  getPortalSession(): Observable<StripeUrlModel> {
    return this.http.get<StripeUrlModel>(`${API_SUBSCRIPTION_URL}/customer-portal`).pipe(
      catchError((err) => {
        console.error('Failed to create portal session', err);
        return throwError(() => new Error('Portal session failed'));
      })
    );
  }

  subscribeWithPromoCode(code: string): Observable<any> {
    this.isLoading$.next(true);
    return this.http.post(`${API_SUBSCRIPTION_URL}/promo-code?promo-code=${code}`, {}).pipe(
      finalize(() => this.isLoading$.next(false)),
      tap(() => this.fetchSubscription().subscribe()),
      catchError((err) => {
        console.error('Failed to subscribe with promo code', err);
        return throwError(err);
      })
    );
  }

  setBillingEmail(email: string) {
    this.isLoading$.next(true);
    return this.http.put(`${API_SUBSCRIPTION_URL}/billing-email?email=${email}`, {}).pipe(
      finalize(() => this.isLoading$.next(false))
    );
  }

  validateFeature(feature: string, projectId?: string, versionId?: string): Observable<boolean> {
    let url = `${API_SUBSCRIPTION_URL}/validate-feature?feature=${feature}`;
    if (projectId) {
      url += `&project-id=${projectId}`;
    }
    if (versionId) {
      url += `&version-id=${versionId}`;
    }
    return this.http.get<FeatureValidation>(url).pipe(map(fv => fv.valid));
  }

  refreshSubscription() {
    this.subscription$.next(this.subscription$.value);
  }

  ngOnDestroy() {
    this.subs.forEach((sb) => sb.unsubscribe());
  }
}
