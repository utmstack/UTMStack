import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from 'src/environments/environment';

@Injectable({ providedIn: 'root' })
export class PlanHttpService {
  public resourceUrl = environment.CUSTOMMER_MANAGER_API_URL + 'api/v1/subscription';

  constructor(protected http: HttpClient) {}

  getPlans(): Observable<any> {
    return this.http.get<any>(`${this.resourceUrl}/plans`);
  }

  getSubscription(): Observable<any> {
    return this.http.get<any>(this.resourceUrl);
  }

  getCheckoutSession(price_id:string): Observable<any> {
    return this.http.get<any>(`${this.resourceUrl}/checkout?price_id=${price_id}`);
  }

  getCustomerPortal(): Observable<any> {
    return this.http.get<any>(`${this.resourceUrl}/customer-portal`);
  }

  stripeWebhook(payload: any): Observable<any> {
    return this.http.post<any>(`${this.resourceUrl}/webhook`, payload);
  }

  updateBillingEmail(email: string): Observable<any> {
    return this.http.put<any>(`${this.resourceUrl}/billing-email`, { email });
  }

  validateFeature(): Observable<any> {
    return this.http.get<any>(`${this.resourceUrl}/validate-feature`);
  }
}
