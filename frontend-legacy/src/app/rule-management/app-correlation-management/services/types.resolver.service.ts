import {HttpResponse} from '@angular/common/http';
import { Injectable } from '@angular/core';
import {ActivatedRouteSnapshot, Resolve, RouterStateSnapshot} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {Observable} from 'rxjs';
import {tap} from 'rxjs/operators';
import { DataType } from '../../models/rule.model';
import {DataTypeService} from '../../services/data-type.service';

export const itemsPerPage = 25;

@Injectable()
export class TypesResolverService implements Resolve<Observable<HttpResponse<DataType[]>>> {

    constructor(private typeService: DataTypeService,
                private spinner: NgxSpinnerService) {}
    resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<HttpResponse<DataType[]>> {
        this.spinner.show('loadingSpinner');
        return this.typeService.getAll({page: 0, size: itemsPerPage})
            .pipe(tap(() => this.spinner.hide('loadingSpinner')));
    }
}
