import {HttpEvent, HttpHandler, HttpInterceptor, HttpRequest} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../app.constants';
import {FederationInstanceStateService} from '../../federation/services/federation-instance-state.service';
import {FederationModeService} from '../../federation/services/federation-mode.service';

const AUTH_PATH_PREFIXES = [
  'api/authenticate',
  'api/v1/auth/',
  'api/tfa/',
  'auth/'
];

@Injectable()
export class FederationInstanceInterceptor implements HttpInterceptor {
  constructor(private modeService: FederationModeService,
              private instanceState: FederationInstanceStateService) {}

  intercept(request: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    if (!this.modeService.isActive) {
      return next.handle(request);
    }
    if (!this.isApiRequest(request.url)) {
      return next.handle(request);
    }
    if (this.isAuthEndpoint(request.url)) {
      return next.handle(request);
    }
    const instance = this.instanceState.current;
    if (!instance) {
      return next.handle(request);
    }
    const cloned = request.clone({
      setHeaders: {'X-UTM-Instance': String(instance.id)}
    });
    return next.handle(cloned);
  }

  private isApiRequest(url: string): boolean {
    if (!url) {
      return false;
    }
    if (/^https?:/i.test(url)) {
      return !!SERVER_API_URL && url.startsWith(SERVER_API_URL);
    }
    return true;
  }

  private isAuthEndpoint(url: string): boolean {
    return AUTH_PATH_PREFIXES.some(prefix => url.includes(prefix));
  }
}
