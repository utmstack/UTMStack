import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {HippaSignaturesType} from '../type/hippa-signatures.type';


@Injectable({
  providedIn: 'root'
})
export class ComplianceTemplateService {
  private resourceUrl = SERVER_API_URL + 'api/compliance-templates';

  constructor(private http: HttpClient) {
  }

  getHIPPASignatures(): Observable<HttpResponse<HippaSignaturesType[]>> {
    return this.http.get<HippaSignaturesType[]>(SERVER_API_URL + 'api/compliance/hipaa/tools-version-and-signature',
      {observe: 'response'});
  }
}
