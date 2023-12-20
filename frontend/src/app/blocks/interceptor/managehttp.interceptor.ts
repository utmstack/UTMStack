import {HttpEvent, HttpHandler, HttpInterceptor, HttpRequest} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {NavigationEnd, Router} from '@angular/router';
import {Observable} from 'rxjs';
import {takeUntil} from 'rxjs/operators';
import {BYPASS_ROUTES} from '../../shared/constants/app-routes.constant';
import {HttpCancelService} from '../service/httpcancel.service';


@Injectable()
export class ManageHttpInterceptor implements HttpInterceptor {

  constructor(router: Router,
              private httpCancelService: HttpCancelService) {
    router.events.subscribe(event => {
      // An event triggered at the end of the activation part of the Resolve phase of routing
      // and is not a notification request
      if (event instanceof NavigationEnd) {
        // Cancel pending callsA
        this.httpCancelService.cancelPendingRequests();
      }
    });
  }

  /**
   * Determine if cancel request or not on page request
   * @param route Current route
   */
  filterRequest(route: string): boolean {
    return BYPASS_ROUTES.findIndex(value => route.includes(value)) !== -1;
  }

  intercept<T>(req: HttpRequest<T>, next: HttpHandler): Observable<HttpEvent<T>> {
    if (!this.filterRequest(req.url)) {
      return next.handle(req).pipe(takeUntil(this.httpCancelService.onCancelPendingRequests()));
    } else {
      return next.handle(req);
    }
  }
}
