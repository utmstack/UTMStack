import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';

export enum TfaStage {
  INIT,
  VERIFY,
  COMPLETE
}

export enum TfaMethod {
  EMAIL = 'EMAIL',
  TOTP = 'TOTP'
}

export interface TfaInitRequest {
  method: TfaMethod;
}

export interface TfaInitResponse {
  method: string;
  delivery: {
    target?: string;
    code?: string;
  };
  expiresInSeconds: number;
}

export interface TfaVerifyRequest {
  method: TfaMethod;
  code: string;
}

export interface TfaSaveRequest {
  method: TfaMethod;
  enable: boolean;
}

export interface TfaVerifyResponse {
  valid: boolean;
  expired: boolean;
  message?: string;
  expiresInSeconds?: number;
}

export interface TfaEnrollRequest {
  initMethod: TfaMethod;
  verifyMethod?: TfaMethod;
  verifyCode?: string;
  completeRequest?: TfaSaveRequest;
}

@Injectable({
  providedIn: 'root'
})
export class TfaService {
  private readonly baseUrl = `${SERVER_API_URL}api/tfa`;
  private readonly enrollBaseUrl = `${SERVER_API_URL}api/enrollment/tfa`;

  constructor(private http: HttpClient) {}

  initTfa(request: TfaInitRequest): Observable<TfaInitResponse> {
    return this.http.post<TfaInitResponse>(`${this.baseUrl}/init`, request);
  }

  verifyTfa(request: TfaVerifyRequest): Observable<TfaVerifyResponse> {
    return this.http.post<TfaVerifyResponse>(`${this.baseUrl}/verify`, request);
  }

  completeTfa(request: TfaSaveRequest): Observable<TfaVerifyResponse> {
    return this.http.post<TfaVerifyResponse>(`${this.baseUrl}/complete`, request);
  }


  enrollTfa(request: any): Observable<any> {
    return this.http.post<TfaVerifyResponse>(`${this.enrollBaseUrl}`, request);
  }


}
