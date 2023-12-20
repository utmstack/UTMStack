import {Injectable} from '@angular/core';
import {ActivatedRouteSnapshot, CanDeactivate, RouterStateSnapshot} from '@angular/router';
import {Observable} from 'rxjs';
import {AppConfigSectionsComponent} from '../../app-config-sections/app-config-sections.component';
import {DialogService} from '../services/dialog-service';

@Injectable()
export class CanDeactivateConfigGuard implements CanDeactivate<AppConfigSectionsComponent> {

  constructor(private dialogService: DialogService) {
  }

  canDeactivate(
    component: AppConfigSectionsComponent,
    route: ActivatedRouteSnapshot,
    state: RouterStateSnapshot
  ): Observable<boolean> | boolean {
    const url: string = state.url;
    if (component.configToSave.length > 0) {
      return this.dialogService.confirm('Discard changes for this configuration?');
    }
    return true;
  }
}
