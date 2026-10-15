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
    this.printBlob(url)
  }

 printBlob(blobUrl) {
  const iframe = document.createElement('iframe');

  iframe.style.position = 'fixed';
  iframe.style.right = '0';
  iframe.style.bottom = '0';
  iframe.style.width = '0';
  iframe.style.height = '0';
  iframe.style.border = 'none';

  iframe.src = blobUrl;

  document.body.appendChild(iframe);

  iframe.onload = function() {
    iframe.contentWindow.focus();
    iframe.contentWindow.print();

    setTimeout(() => {
      document.body.removeChild(iframe);
    }, 1000);   };
}


}
