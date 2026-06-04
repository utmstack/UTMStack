import {HttpErrorResponse, HttpEvent, HttpHandler, HttpInterceptor, HttpRequest} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {tap} from 'rxjs/operators';
import {AuthServerProvider} from '../../core/auth/auth-jwt.service';
import {HttpCancelService} from '../service/httpcancel.service';
import {AccountService} from "../../core/auth/account.service";

@Injectable()
export class AuthExpiredInterceptor implements HttpInterceptor {
  constructor(private authServerProvider: AuthServerProvider,
              private httpCancelService: HttpCancelService,
              private accountService: AccountService) {
  }

  intercept(request: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    return next.handle(request).pipe(
      tap(
        (event: HttpEvent<any>) => {
        },
        (err: any) => {
          if (err instanceof HttpErrorResponse) {
            if (err.status === 401 || err.status === 403) {
              this.authServerProvider.logout().subscribe(() => {
                this.accountService.authenticate(null);
                this.httpCancelService.cancelPendingRequests();
                console.log('UTMStack 401');
              });
            }
          }
        }
      )
    );
  }
}
