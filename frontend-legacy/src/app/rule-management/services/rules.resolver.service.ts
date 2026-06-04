import {HttpResponse} from '@angular/common/http';
import { Injectable } from '@angular/core';
import {ActivatedRouteSnapshot, Resolve, RouterStateSnapshot} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {Observable} from 'rxjs';
import {tap} from 'rxjs/operators';
import {Rule, RULE_REQUEST} from '../models/rule.model';
import {RuleService} from './rule.service';

@Injectable()
export class RulesResolverService implements Resolve<HttpResponse<Rule[]>> {

    constructor(private ruleService: RuleService,
                private spinner: NgxSpinnerService) {}
    resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<HttpResponse<Rule[]>> {
        this.spinner.show('loadingSpinner');
        return this.ruleService.fetchData(RULE_REQUEST)
            .pipe(tap(() => this.spinner.hide('loadingSpinner')));
    }
}
