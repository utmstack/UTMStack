import {Injectable} from '@angular/core';
import {ActivatedRouteSnapshot, CanActivate, Router, RouterStateSnapshot} from '@angular/router';
import {UtmModulesEnum} from '../../app-module/shared/enum/utm-module.enum';
import {UtmModulesService} from '../../app-module/shared/services/utm-modules.service';
import {ActiveAdModuleActiveService} from '../../shared/services/active-modules/active-ad-module.service';

@Injectable({providedIn: 'root'})
export class ActiveModuleRouteAccessService implements CanActivate {
  constructor(
    private router: Router,
    private adModuleActiveService: ActiveAdModuleActiveService,
    private modulesService: UtmModulesService
  ) {
  }

  canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): boolean | Promise<boolean> {
    const module = route.data.module;
    return this.checkActiveModule(module, state.url);
  }

  checkActiveModule(module: UtmModulesEnum, url: string): Promise<boolean> {
    return new Promise<boolean>(resolve => {
      this.modulesService.isActive(module).subscribe(response => {
        if (!response.body) {
          this.router.navigate(['/data-sources/sources'], {
            queryParams: {setUp: module}
          });
        }
        resolve(response.body);
      });
    });
  }
}
