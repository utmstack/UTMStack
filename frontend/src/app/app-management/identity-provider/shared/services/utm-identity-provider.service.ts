import { HttpClient, HttpResponse } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import {SERVER_API_URL} from '../../../../app.constants';
import {createRequestOption} from '../../../../shared/util/request-util';
import { UtmIdentityProvider } from '../models/utm-identity-provider.model';

@Injectable({
  providedIn: 'root'
})
export class UtmIdentityProviderService {
  private resourceUrl = SERVER_API_URL + 'api/identity-providers';

  constructor(private http: HttpClient) {}

  create(provider: UtmIdentityProvider, privateKey?: File | null, certificate?: File | null)
    : Observable<HttpResponse<UtmIdentityProvider>> {

    const formData = this.convertToFormData(provider, privateKey, certificate);
    return this.http.post<UtmIdentityProvider>(this.resourceUrl, formData, { observe: 'response' });
  }

  update(id: number, provider: UtmIdentityProvider, privateKey?: File | null, certificate?: File | null):
    Observable<HttpResponse<UtmIdentityProvider>> {

    const formData = this.convertToFormData(provider, privateKey, certificate);
    return this.http.put<UtmIdentityProvider>(`${this.resourceUrl}/${id}`, formData, { observe: 'response' });
  }

  find(id: number): Observable<HttpResponse<UtmIdentityProvider>> {
    return this.http.get<UtmIdentityProvider>(`${this.resourceUrl}/${id}`, { observe: 'response' });
  }

  query(request: any): Observable<HttpResponse<UtmIdentityProvider[]>> {
    const params = createRequestOption(request);
    return this.http.get<UtmIdentityProvider[]>(this.resourceUrl, { params, observe: 'response' });
  }

  delete(id: number): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${id}`, { observe: 'response' });
  }

  testConnection(provider: UtmIdentityProvider): Observable<HttpResponse<{ success: boolean; message: string }>> {
    return this.http.post<{ success: boolean; message: string }>(`${this.resourceUrl}/test`, provider, { observe: 'response' });
  }

  private convertToFormData(data: any, privateKey?: File | null, certificate?: File | null): FormData {
    const formData = new FormData();

    formData.append('name', data.name);
    formData.append('providerType', data.providerType);
    formData.append('metadataUrl', data.metadataUrl);
    formData.append('active', data.active.toString());

    if (privateKey) {
      formData.append('spPrivateKeyFile', privateKey);
    }
    if (certificate) {
      formData.append('spCertificateFile', certificate);
    }

    return formData;
  }
}
