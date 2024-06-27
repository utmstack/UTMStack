import { Injectable } from '@angular/core';
import {ActivatedRouteSnapshot, Resolve, RouterStateSnapshot} from '@angular/router';
import {Observable} from 'rxjs';
import { Rule } from '../models/rule.model';
import {RuleService} from './rule.service';


@Injectable()
export class RulesResolverService implements Resolve<Observable<Rule[]>> {

    constructor(private ruleService: RuleService) {}
    resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<Rule[]> {
        return this.ruleService.loadAll();
    }
}
