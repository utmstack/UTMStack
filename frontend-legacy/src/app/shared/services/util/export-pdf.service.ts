import { HttpClient, HttpResponse } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';

@Injectable({
  providedIn: 'root'
})
export class ExportPdfService {

  public resourceUrl = SERVER_API_URL + 'api';

  constructor(private http: HttpClient) { }

  getPdf(url: string, filename: string, accessType: string): Observable<HttpResponse<Blob>> {
    const params = `?url=${encodeURIComponent(url)}&filename=${encodeURIComponent(filename)}&accessType=${accessType}`;
    const urlWithParams = `${this.resourceUrl}/generate-pdf-report${params}`;

    return this.http.get(`${urlWithParams}`, {
      observe: 'response',
      responseType: 'blob'
    });
  }

  handlePdfResponse(response: any): void {
    const blob = new Blob([response.body], { type: 'application/pdf' });
    const url = window.URL.createObjectURL(blob);
    window.open(url, '_blank');
  }
}
