import {Component, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import {ResizeEvent} from 'angular-resizable-element';
import {Observable} from 'rxjs';
import {map, tap} from 'rxjs/operators';
import {EventDataTypeEnum} from '../../data-management/alert-management/shared/enums/event-data-type.enum';
import {AddRuleComponent} from '../components/add-rule/add-rule.component';
import {Rule} from '../models/rule.model';

@Component({
    selector: 'app-rules',
    templateUrl: './app-rule.component.html',
    styleUrls: ['./app-rule.component.css']
})
export class AppRuleComponent implements OnInit {

    dataType: EventDataTypeEnum = EventDataTypeEnum.ALERT;
    rules$: Observable<Rule[]>;

    constructor(private route: ActivatedRoute,
                private modalService: NgbModal) {
    }

    ngOnInit() {
        this.rules$ = this.route.data.pipe(
            tap((data: { rules: Rule[] }) => {
                console.log(data.rules);
            }),
            map((data: { rules: Rule[] }) => data.rules)
        );
    }

    addRule() {

        const modalGroup = this.modalService.open(AddRuleComponent, {centered: true});

    }

    onResize($event: ResizeEvent) {

    }

    loadPage($event: number) {

    }

    onItemsPerPageChange($event: number) {

    }
}
