import {HttpResponse} from '@angular/common/http';
import { Injectable } from '@angular/core';
import {ActivatedRouteSnapshot, Resolve, RouterStateSnapshot} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {Observable} from 'rxjs';
import {tap} from 'rxjs/operators';
import {itemsPerPage, RegexPattern} from '../../models/rule.model';
import {PatternManagerService} from './pattern-manager.service';

@Injectable()
export class PatternsResolverService implements Resolve<Observable<HttpResponse<RegexPattern[]>>> {

    constructor(private patternManagerService: PatternManagerService,
                private spinner: NgxSpinnerService) {}
    resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<HttpResponse<RegexPattern[]>> {
        this.spinner.show('loadingSpinner');
        return this.patternManagerService.getAll({page: 0, size: itemsPerPage})
            .pipe(tap(() => this.spinner.hide('loadingSpinner')));
    }
}
