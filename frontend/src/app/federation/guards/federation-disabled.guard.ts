import {Injectable} from '@angular/core';
import {CanActivate, CanActivateChild, Router} from '@angular/router';
import {FederationModeService} from '../services/federation-mode.service';

@Injectable({providedIn: 'root'})
export class FederationDisabledGuard implements CanActivate, CanActivateChild {
  constructor(private federationModeService: FederationModeService,
              private router: Router) {}

  canActivate(): boolean {
    if (this.federationModeService.isActive) {
      this.router.navigate(['/federation/welcome']);
      return false;
    }
    return true;
  }

  canActivateChild(): boolean {
    return this.canActivate();
  }
}
