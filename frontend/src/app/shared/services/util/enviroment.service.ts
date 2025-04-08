import { Injectable } from '@angular/core';
import {environment} from '../../../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class EnvironmentService {

  constructor() { }

  isDev(): boolean {
    return !environment.production;
  }

  isProd(): boolean {
    return environment.production;
  }
}
