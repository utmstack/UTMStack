import {Component, Input, OnInit} from '@angular/core';
import {ResizeEvent} from 'angular-resizable-element';
import { SortEvent } from 'src/app/shared/directives/sortable/type/sort-event';
import {ALERT_TIMESTAMP_FIELD} from '../../../shared/constants/alert/alert-field.constant';
import {UtmFieldType} from '../../../shared/types/table/utm-field.type';
import {RULE_FIELDS} from '../../models/rule.constant';
import { Rule } from '../../models/rule.model';


@Component({
  selector: 'app-rule-list',
  templateUrl: './rule-list.component.html',
  styleUrls: ['./rule-list.component.css']
})
export class RuleListComponent implements OnInit {

  @Input()
  rules: Rule[];

  sortEvent: SortEvent;
  sortBy = ALERT_TIMESTAMP_FIELD + ',desc';
  fields = RULE_FIELDS;
  checkbox: any;
  rulesSelected: Rule[] = [];

  page: number;
  totalItems: number;
  itemsPerPage: number;

  dataType: any;
  loading: any;
  viewRuleDetail: any;
  ruleDetail: Rule;
  constructor() { }

  ngOnInit() {
  }

  addRule() {

  }

  onResize($event: ResizeEvent) {

  }

  loadPage($event: number) {

  }

  onItemsPerPageChange($event: number) {

  }
  getRowToFiltersData(alert: any) {
    /*const modalRef = this.modalService.open(RowToFiltersComponent, {centered: true});
    modalRef.componentInstance.alert = alert;
    modalRef.componentInstance.fields = this.fields;
    modalRef.componentInstance.addRowToFilter.subscribe(filterRow => {
      mergeParams(filterRow, this.filters).then(value => {
        this.filters = value;
        this.page = 1;
        this.getAlert('on add row to filter');
        this.alertFiltersBehavior.$filters.next(this.filters);
      });
    });*/
  }

  onRefreshData($event: boolean) {

  }

  onSortBy($event: SortEvent) {
    this.sortBy = $event.column + ',' + $event.direction;
    // this.getAlert('on sort by');
  }
  toggleCheck() {
    this.checkbox = !this.checkbox;
    if (!this.checkbox) {
      this.rulesSelected = [];
    } else {
      for (const rule of this.rules) {
        const index = this.rulesSelected.indexOf(rule);
        if (index === -1) {
          this.rulesSelected.push(rule);
        }
      }
    }
  }

  addToSelected(alert: any) {
    const index = this.rulesSelected.indexOf(alert);
    if (index === -1) {
      this.rulesSelected.push(alert);
    } else {
      this.rulesSelected.splice(index, 1);
    }
  }

  isSelected(alert: any): boolean {
    return this.rulesSelected.findIndex(value => value.id === alert.id) !== -1;
  }

  viewDetailAlert(rule: any, td: UtmFieldType) {
    this.ruleDetail = rule;
    this.viewRuleDetail = true;
  }
}
