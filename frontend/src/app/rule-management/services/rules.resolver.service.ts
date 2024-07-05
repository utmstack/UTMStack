import {HttpResponse} from '@angular/common/http';
import { Injectable } from '@angular/core';
import {ActivatedRouteSnapshot, Resolve, RouterStateSnapshot} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {Observable} from 'rxjs';
import {tap} from 'rxjs/operators';
import { Rule } from '../models/rule.model';
import {RuleService} from './rule.service';

export const itemsPerPage = 10;

@Injectable()
export class RulesResolverService implements Resolve<Observable<HttpResponse<Rule[]>>> {

    constructor(private ruleService: RuleService,
                private spinner: NgxSpinnerService) {}
    resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<HttpResponse<Rule[]>> {
        this.spinner.show('loadingSpinner');
        return this.ruleService.getRules({page: 0, size: itemsPerPage})
            .pipe(tap(() => this.spinner.hide('loadingSpinner')));
    }
}
