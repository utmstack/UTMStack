import {Component, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import {ResizeEvent} from 'angular-resizable-element';
import {Observable} from 'rxjs';
import {map, tap} from 'rxjs/operators';
import {EventDataTypeEnum} from '../../data-management/alert-management/shared/enums/event-data-type.enum';
import {FILTER_RULE_FIELDS, RuleFilterType} from '../models/rule.constant';
import {Rule} from '../models/rule.model';

import {HttpResponse} from "@angular/common/http";
import {FilterService} from "../services/filter.service";
import {AssetFieldFilterEnum} from "../../assets-discover/shared/enums/asset-field-filter.enum";
import {SERVER_API_URL} from "../../app.constants";
import {AssetFilterType} from "../../assets-discover/shared/types/asset-filter.type";
import {ITEMS_PER_PAGE} from "../../shared/constants/pagination.constants";


@Component({
    selector: 'app-rules',
    templateUrl: './app-rule.component.html',
    styleUrls: ['./app-rule.component.css']
})
export class AppRuleComponent implements OnInit {

    constructor(private filterService: FilterService) {
    }

    dataType: EventDataTypeEnum = EventDataTypeEnum.ALERT;
    fieldFilters = FILTER_RULE_FIELDS;
    filterUrl = `${SERVER_API_URL}api/correlation-rule/search-property-values`;
    requestParam = {
        page: 0,
        size: ITEMS_PER_PAGE,
        sort: 'id,desc',
    };

    ngOnInit() {}

    addRule() {

        // const modalGroup = this.modalService.open(AddRuleComponent, {centered: true});

    }

    onResize($event: ResizeEvent) {

    }

    loadPage($event: number) {

    }

    onItemsPerPageChange($event: number) {

    }


    resetAllFilters() {
        this.filterService.resetAllFilters();
    }

    /*onFilterChange($event: { prop: string, values: any }) {
        this.requestParam = {
            ...this.requestParam,
            [$event.prop]: $event.values.length > 0 ? $event.values : null
        };
        console.log(this.requestParam);

       /!* this.assetFiltersBehavior.$assetAppliedFilter.next(this.requestParam);
        this.assetFiltersBehavior.$assetFilter.next(this.requestParam);*!/
        this.getAssets();
    }*/
}
