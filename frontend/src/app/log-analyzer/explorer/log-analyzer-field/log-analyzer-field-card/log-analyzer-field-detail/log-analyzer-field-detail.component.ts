import {Component, Input, OnDestroy, OnInit} from '@angular/core';
import {Observable} from 'rxjs';
import {UtmFilterBehavior} from '../../../../../shared/components/utm/filters/utm-elastic-filter/shared/behavior/utm-filter.behavior';
import {LOG_ANALYZER_TOTAL_ITEMS} from '../../../../../shared/constants/log-analyzer.constant';
import {ElasticDataTypesEnum} from '../../../../../shared/enums/elastic-data-types.enum';
import {ElasticOperatorsEnum} from '../../../../../shared/enums/elastic-operators.enum';
import {ElasticSearchFieldInfoType} from '../../../../../shared/types/elasticsearch/elastic-search-field-info.type';
import {ElasticFilterType} from '../../../../../shared/types/filter/elastic-filter.type';
import {DataSortBehavior} from '../../../../shared/behaviors/data-sort.behavior';
import {LogFilterBehavior} from '../../../../shared/behaviors/log-filter.behavior';
import {LogAnalyzerService} from '../../../../shared/services/log-analyzer.service';
import {LogAnalyzerFieldDetailType} from '../../../../shared/type/log-analyzer-field-detail.type';

@Component({
  selector: 'app-log-analyzer-field-detail',
  templateUrl: './log-analyzer-field-detail.component.html',
  styleUrls: ['./log-analyzer-field-detail.component.scss']
})
export class LogAnalyzerFieldDetailComponent implements OnInit, OnDestroy {
  @Input() field: ElasticSearchFieldInfoType;
  @Input() pattern: string;
  filters: ElasticFilterType[] = [];
  fieldTopValues: LogAnalyzerFieldDetailType;
  loading = true;
  total = LOG_ANALYZER_TOTAL_ITEMS;
  public logObservable: Observable<{ filter: ElasticFilterType[], sort: string }>;
  subscription: any;
  private sort = '@timestamp,desc';

  constructor(private logAnalyzerService: LogAnalyzerService,
              private logFilterBehavior: LogFilterBehavior,
              private dataSortBehavior: DataSortBehavior,
              private utmFilterBehavior: UtmFilterBehavior) {
  }

  ngOnInit() {
    this.logObservable = this.logFilterBehavior.$logFilter.asObservable();
    this.subscription = this.logObservable.subscribe(filterValue => {
      if (filterValue) {
        this.filters = filterValue.filter;
        this.sort = filterValue.sort;
        this.getFieldTopValue();
      } else {
        this.getFieldTopValue();
      }
    });
  }

  ngOnDestroy(): void {
    this.subscription.unsubscribe();
  }

  getFieldTopValue() {
    this.logAnalyzerService.getFieldTopValues(this.pattern,
      (((this.field.type === ElasticDataTypesEnum.TEXT ||
        this.field.type === ElasticDataTypesEnum.STRING) && !this.field.name.includes('.keyword'))
        ? this.field.name + '.keyword' : this.field.name),
      5, this.filters, this.sort).subscribe(value => {
      this.fieldTopValues = value;
      this.loading = false;
    });
  }


  toggleFilter(value: string) {
    this.utmFilterBehavior.$filterChange.next({
      operator: ElasticOperatorsEnum.IS,
      field: this.field.name,
      value,
      status: 'ACTIVE'
    });
  }
}
