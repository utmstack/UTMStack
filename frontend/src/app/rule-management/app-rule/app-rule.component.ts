import {Component, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import {ResizeEvent} from 'angular-resizable-element';
import {Observable} from 'rxjs';
import {map, tap} from 'rxjs/operators';
import {EventDataTypeEnum} from '../../data-management/alert-management/shared/enums/event-data-type.enum';
import {FILTER_RULE_FIELDS} from '../models/rule.constant';
import {Rule} from '../models/rule.model';
import {AddRuleComponent} from './components/add-rule/add-rule.component';


@Component({
    selector: 'app-rules',
    templateUrl: './app-rule.component.html',
    styleUrls: ['./app-rule.component.css']
})
export class AppRuleComponent implements OnInit {

    constructor(private route: ActivatedRoute,
                private modalService: NgbModal) {
    }

    dataType: EventDataTypeEnum = EventDataTypeEnum.ALERT;
    rules$: Observable<Rule[]>;
    fieldFilters = FILTER_RULE_FIELDS;


    ngOnInit() {
        this.rules$ = this.route.data.pipe(
            tap((data: { rules: Rule[] }) => {
                console.log(data.rules);
            }),
            map((data: { rules: Rule[] }) => data.rules)
        );
    }

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

    }
}
