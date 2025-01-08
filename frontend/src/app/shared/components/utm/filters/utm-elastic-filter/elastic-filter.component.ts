import {Component, EventEmitter, Input, OnInit, Output, ViewChild} from '@angular/core';
import {NgbActiveModal, NgbModal, NgbPopover} from '@ng-bootstrap/ng-bootstrap';
import {FILTER_OPERATORS} from '../../../../constants/filter-operators.const';
import {ElasticOperatorsEnum} from '../../../../enums/elastic-operators.enum';
import {NatureDataPrefixEnum} from '../../../../enums/nature-data.enum';
import {ElasticFilterType} from '../../../../types/filter/elastic-filter.type';
import {OperatorsType} from '../../../../types/filter/operators.type';
import {TimeFilterType} from '../../../../types/time-filter.type';
import {ElasticFilterDefaultTime} from '../elastic-filter-time/elastic-filter-time.component';
import {UtmFilterBehavior} from './shared/behavior/utm-filter.behavior';

@Component({
  selector: 'app-utm-elastic-filter',
  templateUrl: './elastic-filter.component.html',
  styleUrls: ['./elastic-filter.component.scss']
})
export class ElasticFilterComponent implements OnInit {
  @Output() filterChange = new EventEmitter<ElasticFilterType[]>();
  @Input() pattern: string;
  @Input() filters: ElasticFilterType[] = [];
  @Input() defaultTime: ElasticFilterDefaultTime;
  operators: OperatorsType[] = FILTER_OPERATORS;
  operatorEnum = ElasticOperatorsEnum;
  @ViewChild('popoverFilter') popoverFilter: NgbPopover;
  filterSelected: ElasticFilterType;
  indexEdit: number;
  editMode: boolean;

  constructor(public modalService: NgbModal,
              private activeModal: NgbActiveModal,
              private utmFilterBehavior: UtmFilterBehavior) {
  }

  ngOnInit() {
    this.filters = this.filters ? this.filters : [];
    this.utmFilterBehavior.$filterChange.subscribe(filter => {
      if (filter) {
        this.filters.push(filter);
        this.filterChange.emit(this.filters);
      }
    });
  }

  addFilter($event: ElasticFilterType) {
    this.popoverFilter.close();
    if (!this.editMode) {
      this.filters.push($event);
    } else {
      this.filters[this.indexEdit] = $event;
    }
    this.editMode = false;
    this.filterSelected = null;
    this.filterChange.emit(this.filters);
  }

  onTimeFilterChange($event: TimeFilterType) {
    const indexTime = this.filters.findIndex(value =>
      value.field === NatureDataPrefixEnum.TIMESTAMP && value.operator === this.operatorEnum.IS_BETWEEN);
    const filter = {
      field: NatureDataPrefixEnum.TIMESTAMP,
      operator: this.operatorEnum.IS_BETWEEN,
      value: [$event.timeFrom, $event.timeTo]
    };
    if (indexTime === -1) {
      this.filters.push(filter);
    } else {
      this.filters[indexTime].value = [$event.timeFrom, $event.timeTo];
    }
    this.filterChange.emit(this.filters);
  }

  deleteFilter(filter: ElasticFilterType) {
    const index = this.filters.findIndex(value => {
      return (value.field === filter.field && value.operator === filter.operator);
    });
    this.filters.splice(index, 1);
    this.filterSelected = undefined;
    this.filterChange.emit(this.filters);
  }

  deleteAll() {
    this.filters = [];
    this.filterSelected = undefined;
    this.filterChange.emit(this.filters);
  }

  extractOperator(operator: string): string {
    const index = this.operators.findIndex(value => value.operator === operator);
    return this.operators[index].name;
  }

  getFilterLabel(filter: ElasticFilterType): string {
    return filter.field + ' ' +
      this.extractOperator(filter.operator) + ' ' +
      (filter.value ? filter.value.toString().replace(',', ' and ') : '');
  }

  invertAction() {
    const filter = this.filters[this.filters.indexOf(this.filterSelected)];
    const operator: OperatorsType = this.operators[this.operators.findIndex(value => value.operator === filter.operator)];
    filter.operator = operator.inverse;
    this.filterChange.emit(this.filters);
  }

  resolveFilters(): ElasticFilterType[] {
    return this.filters.filter(value => value.operator !== ElasticOperatorsEnum.IS_IN_FIELD);
  }
}
