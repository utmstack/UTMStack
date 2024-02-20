import {Injectable} from '@angular/core';
import {AccountService} from '../auth/account.service';
import {AuthServerProvider} from '../auth/auth-jwt.service';


@Injectable({providedIn: 'root'})
export class LoginService {

  constructor(private accountService: AccountService,
              private authServerProvider: AuthServerProvider) {
  }

  login(credentials): Promise<any> {
    return new Promise((resolve, reject) => {
      this.authServerProvider.login(credentials).subscribe(
        data => {
          if (data) {
            this.accountService.identity(true).then(account => {
              resolve(data);
            });
          } else {
            resolve(data);
          }
        },
        err => {
          reject(err);
        }
      );
    });
  }

  loginWithToken(jwt, rememberMe) {
    return this.authServerProvider.loginWithToken(jwt, rememberMe);
  }

  loginWithKey(key, rememberMe) {
    return this.authServerProvider.loginWithAccessKey(key, rememberMe);
  }

  logout() {
    this.authServerProvider.logout().subscribe(() => {
      this.accountService.authenticate(null);
    });
  }

}
