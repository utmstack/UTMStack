import {HttpResponse} from '@angular/common/http';
import { Injectable } from '@angular/core';
import {ActivatedRouteSnapshot, Resolve, RouterStateSnapshot} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {Observable} from 'rxjs';
import {tap} from 'rxjs/operators';
import {Asset, itemsPerPage} from '../../models/rule.model';
import {AssetManagerService} from './asset-manager.service';

@Injectable()
export class AssetsResolverService implements Resolve<Observable<HttpResponse<Asset[]>>> {

    constructor(private assetManagerService: AssetManagerService,
                private spinner: NgxSpinnerService) {}
    resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<HttpResponse<Asset[]>> {
        this.spinner.show('loadingSpinner');
        return this.assetManagerService.getAll({page: 0, size: itemsPerPage})
            .pipe(tap(() => this.spinner.hide('loadingSpinner')));
    }
}
