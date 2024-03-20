import {Component, Input, OnChanges, OnDestroy, OnInit, SimpleChanges} from '@angular/core';
import {Subject} from "rxjs";
import {filter, takeUntil} from 'rxjs/operators';
import {DashboardBehavior} from '../../../../../behaviors/dashboard.behavior';
import {ElasticOperatorsEnum} from '../../../../../enums/elastic-operators.enum';
import {ElasticSearchIndexService} from '../../../../../services/elasticsearch/elasticsearch-index.service';
import {DashboardFilterType} from '../../../../../types/filter/dashboard-filter.type';
import {ElasticFilterType} from '../../../../../types/filter/elastic-filter.type';
import {ChangeFilterValueService} from '../../services/change-filter-value.service';

@Component({
  selector: 'app-dashboard-filter-select',
  templateUrl: './dashboard-filter-select.component.html',
  styleUrls: ['./dashboard-filter-select.component.css']
})
export class DashboardFilterSelectComponent implements OnInit, OnChanges, OnDestroy {
  @Input() filter: DashboardFilterType;
  @Input() fullWidth = false;
  values: any[];
  isLoading = true;
  model = null;
  destroy$: Subject<void> = new Subject<void>();

  constructor(private elasticSearchIndexService: ElasticSearchIndexService,
              private changeFilterValueService: ChangeFilterValueService,
              private dashboardBehavior: DashboardBehavior) {
  }

  ngOnInit() {
    this.getFieldValues();
    this.changeFilterValueService.selectedValue$
      .pipe(
        takeUntil(this.destroy$),
        filter(res => res && this.filter.field === res.field))
      .subscribe(res => this.model = res.value);
  }

  ngOnChanges(changes: SimpleChanges) {
    if (changes && changes.filter) {
      this.getFieldValues();
    }
  }

  getFieldValues() {
    const req = {
      page: 0,
      size: 10,
      indexPattern: this.filter.indexPattern,
      keyword: this.filter.field
    };
    this.isLoading = true;
    this.elasticSearchIndexService.getElasticFieldValues(req).subscribe(res => {
      this.values = res.body;
      this.isLoading = false;
    });
  }

  select($event: any, filter: DashboardFilterType) {
    const elasticFilter: ElasticFilterType = {
      field: filter.field,
      operator: filter.multiple ? ElasticOperatorsEnum.IS_ONE_OF : ElasticOperatorsEnum.IS,
      value: $event
    };
    this.dashboardBehavior.$filterDashboard.next({filter: [elasticFilter], indexPattern: filter.indexPattern});
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
