import { Injectable } from '@angular/core';
import {ActivatedRouteSnapshot, Resolve, RouterStateSnapshot} from '@angular/router';
import {Observable} from 'rxjs';
import { Rule } from '../models/rule.model';
import {RuleService} from './rule.service';
import {NgxSpinnerService} from "ngx-spinner";
import {tap} from "rxjs/operators";
import {HttpResponse} from "@angular/common/http";


@Injectable()
export class RulesResolverService implements Resolve<Observable<HttpResponse<Rule[]>>> {

    constructor(private ruleService: RuleService,
                private spinner: NgxSpinnerService) {}
    resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<HttpResponse<Rule[]>> {
        this.spinner.show('loadingSpinner');
        return this.ruleService.getRules({})
            .pipe(tap(() => this.spinner.hide('loadingSpinner')));
    }
}
