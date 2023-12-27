import {Component, Input, OnChanges, OnDestroy, OnInit, SimpleChanges} from '@angular/core';
import {DashboardBehavior} from '../../../../../behaviors/dashboard.behavior';
import {ElasticOperatorsEnum} from '../../../../../enums/elastic-operators.enum';
import {ElasticSearchIndexService} from '../../../../../services/elasticsearch/elasticsearch-index.service';
import {DashboardFilterType} from '../../../../../types/filter/dashboard-filter.type';
import {ElasticFilterType} from '../../../../../types/filter/elastic-filter.type';
import {ChangeFilterValueService} from '../../services/change-filter-value.service';
import {Subject} from 'rxjs';
import {takeUntil} from 'rxjs/operators';

@Component({
  selector: 'app-dashboard-filter-select',
  templateUrl: './dashboard-filter-select.component.html',
  styleUrls: ['./dashboard-filter-select.component.css']
})
export class DashboardFilterSelectComponent implements OnInit, OnChanges, OnDestroy {
  @Input() filter: DashboardFilterType;
  values: any[];
  isLoading = true;
  selected: any;
  onDestroy$: Subject<void> = new Subject();

  constructor(private elasticSearchIndexService: ElasticSearchIndexService,
              private dashboardBehavior: DashboardBehavior,
              private changeFilterValueService: ChangeFilterValueService) {
  }

  ngOnInit() {
    this.getFieldValues();

    this.changeFilterValueService.selectedValue$
      .pipe(takeUntil(this.onDestroy$))
      .subscribe( data => {
        if (data && this.filter.field === data.field) {
          this.selected = data.value;
        }
      });
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

  ngOnDestroy() {
    this.onDestroy$.next();
    this.onDestroy$.complete();
  }
}
