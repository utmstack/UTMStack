import {Component, Input, OnChanges, OnInit, SimpleChanges} from '@angular/core';
import {DashboardBehavior} from '../../../../../behaviors/dashboard.behavior';
import {ElasticOperatorsEnum} from '../../../../../enums/elastic-operators.enum';
import {ElasticSearchIndexService} from '../../../../../services/elasticsearch/elasticsearch-index.service';
import {DashboardFilterType} from '../../../../../types/filter/dashboard-filter.type';
import {ElasticFilterType} from '../../../../../types/filter/elastic-filter.type';

@Component({
  selector: 'app-dashboard-filter-select',
  templateUrl: './dashboard-filter-select.component.html',
  styleUrls: ['./dashboard-filter-select.component.css']
})
export class DashboardFilterSelectComponent implements OnInit, OnChanges {
  @Input() filter: DashboardFilterType;
  @Input() fullWidth = false;
  values: any[];
  isLoading = true;

  constructor(private elasticSearchIndexService: ElasticSearchIndexService,
              private dashboardBehavior: DashboardBehavior) {
  }

  ngOnInit() {
    this.getFieldValues();
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
}
