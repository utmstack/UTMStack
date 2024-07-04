import {Component, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import {ResizeEvent} from 'angular-resizable-element';
import {Observable} from 'rxjs';
import {map, tap} from 'rxjs/operators';
import {EventDataTypeEnum} from '../../data-management/alert-management/shared/enums/event-data-type.enum';
import {FILTER_RULE_FIELDS} from '../models/rule.constant';
import {Rule} from '../models/rule.model';

import {HttpResponse} from "@angular/common/http";
import {FilterService} from "../services/filter.service";
import {AssetFieldFilterEnum} from "../../assets-discover/shared/enums/asset-field-filter.enum";


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

    onFilterChange($event: { prop: AssetFieldFilterEnum, values: any }) {
        /*switch ($event.prop) {
            case AssetFieldFilterEnum.PORTS:
                this.requestParam.openPorts = $event.values.length > 0 ? $event.values : null;
                break;
            case AssetFieldFilterEnum.SEVERITY:
                this.requestParam.severity = $event.values.length > 0 ? $event.values : null;
                break;
            case AssetFieldFilterEnum.TYPE:
                this.requestParam.type = $event.values.length > 0 ? $event.values : null;
                break;
            case AssetFieldFilterEnum.STATUS:
                this.requestParam.status = $event.values.length > 0 ? $event.values : null;
                break;
            case AssetFieldFilterEnum.ALIAS:
                this.requestParam.alias = $event.values.length > 0 ? $event.values : null;
                break;
            case AssetFieldFilterEnum.OS:
                this.requestParam.os = $event.values.length > 0 ? $event.values : null;
                break;
            case AssetFieldFilterEnum.ALIVE:
                this.requestParam.alive = $event.values;
                break;
            case AssetFieldFilterEnum.PROBE:
                this.requestParam.probe = $event.values.length > 0 ? $event.values : null;
                break;
            case AssetFieldFilterEnum.GROUP:
                this.requestParam.groups = $event.values.length > 0 ? $event.values : null;
                break;
        }
        this.assetFiltersBehavior.$assetAppliedFilter.next(this.requestParam);
        this.assetFiltersBehavior.$assetFilter.next(this.requestParam);
        this.getAssets();*/
    }
}
