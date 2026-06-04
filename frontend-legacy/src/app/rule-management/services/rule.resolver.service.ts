import {HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {ActivatedRouteSnapshot, Resolve, RouterStateSnapshot} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {Observable, of} from 'rxjs';
import {tap} from 'rxjs/operators';
import {Rule} from '../models/rule.model';
import {RuleService} from './rule.service';


@Injectable()
export class RuleResolverService implements Resolve<Observable<HttpResponse<Rule>>> {

    constructor(private ruleService: RuleService,
                private spinner: NgxSpinnerService) {
    }

    resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<HttpResponse<Rule>> {
        this.spinner.show('loadingSpinner');
        const id = parseInt(route.paramMap.get('id'), 10);

        return this.ruleService.getRuleById(id)
            .pipe(
                tap(() => this.spinner.hide('loadingSpinner'))
            );

    }
}
