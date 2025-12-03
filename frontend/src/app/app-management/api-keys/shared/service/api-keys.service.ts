import { Injectable } from '@angular/core';
import { HttpClient, HttpResponse } from '@angular/common/http';
import { Observable } from 'rxjs';
import { SERVER_API_URL } from '../../../../app.constants';
import { ApiKeyResponse } from '../models/ApiKeyResponse';
import { ApiKeyUpsert } from '../models/ApiKeyUpsert';
import {createRequestOption} from "../../../../shared/util/request-util";

/**
 * Service for managing API keys
 */
@Injectable({
  providedIn: 'root'
})
export class ApiKeysService {
  public resourceUrl = SERVER_API_URL + 'api/api-keys';

  constructor(private http: HttpClient) {}

  /**
   * Create a new API key
   */
  create(dto: ApiKeyUpsert): Observable<HttpResponse<ApiKeyResponse>> {
    return this.http.post<ApiKeyResponse>(
      this.resourceUrl,
      dto,
      { observe: 'response' }
    );
  }

  /**
   * Generate (or renew) a plain API key for the given id
   * Returns the plain text key (only once)
   */
  generate(id: string): Observable<HttpResponse<string>> {
    return this.http.post(
      `${this.resourceUrl}/${id}/generate`,
      {},
      { observe: 'response', responseType: 'text' }
    );
  }

  /**
   * Get API key by id
   */
  get(id: string): Observable<HttpResponse<ApiKeyResponse>> {
    return this.http.get<ApiKeyResponse>(
      `${this.resourceUrl}/${id}`,
      { observe: 'response' }
    );
  }

  /**
   * List all API keys (with optional pagination)
   */
  list(params?: any): Observable<HttpResponse<ApiKeyResponse[]>> {
    const httpParams = createRequestOption(params);
    return this.http.get<ApiKeyResponse[]>(
      this.resourceUrl,
      { observe: 'response', params: httpParams },
    );
  }

  /**
   * Update an existing API key
   */
  update(id: string, dto: ApiKeyUpsert): Observable<HttpResponse<ApiKeyResponse>> {
    return this.http.put<ApiKeyResponse>(
      `${this.resourceUrl}/${id}`,
      dto,
      { observe: 'response' }
    );
  }

  /**
   * Delete API key
   */
  delete(id: string): Observable<HttpResponse<void>> {
    return this.http.delete<void>(
      `${this.resourceUrl}/${id}`,
      { observe: 'response' }
    );
  }

  generateApiKey(apiKeyId: string): Observable<HttpResponse<string>> {
    return this.http.post(`${this.resourceUrl}/${apiKeyId}/generate`, null, {
      observe: 'response',
      responseType: 'text'
    });
  }

  /**
   * Search API key usage in Elasticsearch
   */
  usage(params: {
    filters?: any[];
    top: number;
    indexPattern: string;
    includeChildren?: boolean;
    page?: number;
    size?: number;
  }): Observable<any[]> {
    return this.http.get<any[]>(
      `${this.resourceUrl}/usage`,
      {
        params: {
          top: params.top.toString(),
          indexPattern: params.indexPattern,
          includeChildren: params.includeChildren.toString() || 'false',
          page: params.page.toString() || '0',
          size: params.size.toString() || '10'
        }
      }
    );
  }
}

