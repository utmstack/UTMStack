import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';


@Injectable({providedIn: 'root'})
export class LocationService {


  constructor(private http: HttpClient) {
  }


  getCountries(): Observable<any> {
    return this.http.get<any>('assets/data/country.json', {observe: 'response'});
  }

  getStates(): Observable<any> {
    return this.http.get<any>('assets/data/state.json', {observe: 'response'});
  }

}
