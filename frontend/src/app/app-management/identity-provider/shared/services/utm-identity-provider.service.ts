import { HttpClient, HttpResponse } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import {SERVER_API_URL} from '../../../../app.constants';
import { UtmIdentityProvider } from '../models/utm-identity-provider.model';

@Injectable({
  providedIn: 'root'
})
export class UtmIdentityProviderService {
  private resourceUrl = SERVER_API_URL + 'api/identity-providers';

  constructor(private http: HttpClient) {}

  create(provider: UtmIdentityProvider): Observable<HttpResponse<UtmIdentityProvider>> {
    return this.http.post<UtmIdentityProvider>(this.resourceUrl, provider, { observe: 'response' });
  }

  update(provider: UtmIdentityProvider): Observable<HttpResponse<UtmIdentityProvider>> {
    return this.http.put<UtmIdentityProvider>(`${this.resourceUrl}/${provider.id}`, provider, { observe: 'response' });
  }

  find(id: number): Observable<HttpResponse<UtmIdentityProvider>> {
    return this.http.get<UtmIdentityProvider>(`${this.resourceUrl}/${id}`, { observe: 'response' });
  }

  query(): Observable<HttpResponse<UtmIdentityProvider[]>> {
    return this.http.get<UtmIdentityProvider[]>(this.resourceUrl, { observe: 'response' });
  }

  delete(id: number): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${id}`, { observe: 'response' });
  }

  toggleActive(id: number): Observable<HttpResponse<UtmIdentityProvider>> {
    return this.http.patch<UtmIdentityProvider>(`${this.resourceUrl}/${id}/toggle-active`, {}, { observe: 'response' });
  }

  testConnection(provider: UtmIdentityProvider): Observable<HttpResponse<{ success: boolean; message: string }>> {
    return this.http.post<{ success: boolean; message: string }>(`${this.resourceUrl}/test`, provider, { observe: 'response' });
  }
}
