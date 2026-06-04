import {HttpErrorResponse, HttpEvent, HttpHandler, HttpInterceptor, HttpRequest} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {tap} from 'rxjs/operators';
import {AuthServerProvider} from '../../core/auth/auth-jwt.service';
import {HttpCancelService} from '../service/httpcancel.service';

@Injectable()
export class ErrorHandlerInterceptor implements HttpInterceptor {
  constructor(private authServerProvider: AuthServerProvider, private httpCancelService: HttpCancelService) {
  }

  intercept(request: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    return next.handle(request).pipe(
      tap(
        (event: HttpEvent<any>) => {
        },
        (err: any) => {
          if (err instanceof HttpErrorResponse) {
            if (err.status === 502) {
              console.log('UTMStack' + err.status);
            }
          }
        }
      )
    );
  }
}
