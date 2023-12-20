import {HttpResponse} from '@angular/common/http';
import {Component, Input, OnInit} from '@angular/core';
import {Observable} from 'rxjs';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {ChartModel} from '../../../../shared/chart/types/chart.model';
import {UTM_COLOR_THEME} from '../../../../shared/constants/utm-color.const';
import {ChartTypeEnum} from '../../../../shared/enums/chart-type.enum';
import {ElasticDataTypesEnum} from '../../../../shared/enums/elastic-data-types.enum';
import {NatureDataPrefixEnum} from '../../../../shared/enums/nature-data.enum';
import {ElasticSearchIndexService} from '../../../../shared/services/elasticsearch/elasticsearch-index.service';
import {FieldDataService} from '../../../../shared/services/elasticsearch/field-data.service';
import {ElasticSearchFieldInfoType} from '../../../../shared/types/elasticsearch/elastic-search-field-info.type';
import {ElasticFilterType} from '../../../../shared/types/filter/elastic-filter.type';
import {DataNatureBehavior} from '../../behaviors/data-nature.behavior';
import {LogFilterBehavior} from '../../behaviors/log-filter.behavior';
import {LogAnalyzerService} from '../../services/log-analyzer.service';
import {LogAnalyzerChartResponseType} from '../../type/log-analyzer-field-detail.type';

@Component({
  selector: 'app-analyzer-bar-chart',
  templateUrl: './analyzer-bar-chart.component.html',
  styleUrls: ['./analyzer-bar-chart.component.scss']
})
export class AnalyzerBarChartComponent implements OnInit {
  @Input() pattern: string;
  @Input() uuid: string;
  loadingOption = true;
  echartOption: any;
  chart = ChartTypeEnum.BAR_CHART;
  height = '350px';
  data: LogAnalyzerChartResponseType;
  fields: ElasticSearchFieldInfoType[] = [];
  field: ElasticSearchFieldInfoType = {name: NatureDataPrefixEnum.TIMESTAMP, type: ElasticDataTypesEnum.DATE};
  loading = true;
  seriesTypes = ['bar', 'line'];
  chartType = 'bar';
  filters: ElasticFilterType[] = [];
  intervals: string[] = ['Year', 'Quarter', 'Month', 'Week', 'Day', 'Hour', 'Minute', 'Second'];
  interval = 'Day';
  elasticDataTypeEnum = ElasticDataTypesEnum;

  constructor(private elasticSearchIndexService: ElasticSearchIndexService,
              private fieldDataBehavior: FieldDataService,
              private logFilterBehavior: LogFilterBehavior,
              private dataNatureBehavior: DataNatureBehavior,
              private utmToastService: UtmToastService,
              private logAnalyzerService: LogAnalyzerService) {
  }

  ngOnInit() {
    this.getFields();
    this.dataNatureBehavior.$dataNature.subscribe(natureChange => {
      if (natureChange && this.uuid === natureChange.tabUUID) {
        this.pattern = natureChange.pattern.pattern;
        this.getFields();
      }
    });
    this.logFilterBehavior.$logFilter.subscribe(filterValue => {
      if (filterValue.filter) {
        this.filters = filterValue.filter;
        this.getFieldTopValue();
      }
    });
  }

  getFields() {
    this.loading = true;
    this.fieldDataBehavior.getFields(this.pattern).subscribe(field => {
      this.fields = field.filter(value => value.type !== ElasticDataTypesEnum.TEXT || value.name.includes('.keyword'));
      this.field = this.fields[0];
      this.loading = false;
      this.getFieldTopValue();
    });
  }

  chartEvent($event: any) {
    // if (typeof this.visualization.chartAction === 'string') {
    //   this.visualization.chartAction = JSON.parse(this.visualization.chartAction);
    // }
  }

  changeField($event) {
    this.getFieldTopValue();
  }


  changeInterval() {
    this.getFieldTopValue();
  }

  getFieldTopValue() {
    this.echartOption = undefined;
    this.loadingOption = true;
    const req = {
      indexPattern: this.pattern,
      field: this.field.name,
      fieldDataType: this.field.type,
      filters: this.filters,
      interval: this.interval,
      top: 20
    };
    this.logAnalyzerService.chartView(req).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.status, res.headers)
    );
  }

  public buildCharTopFieldValue(data: LogAnalyzerChartResponseType) {
    this.processTopFieldValue(data).subscribe(option => {
      this.echartOption = {
        animation: true,
        color: UTM_COLOR_THEME,
        tooltip: {
          trigger: 'axis',
          backgroundColor: 'rgba(0,75,139,0.85)',
          padding: 10,
          textStyle: {
            fontSize: 13,
            fontFamily: 'Roboto, sans-serif'
          }
        },
        grid: {
          left: '75px',
          top: '35px',
          right: '55px',
          bottom: '60px'
        },
        toolbox: {
          show: true,
          feature: {
            saveAsImage: {
              show: true,
              type: 'png',
              name: 'utm-chart',
              title: 'Save as image'
            },
            dataZoom: {
              show: true,
              title: {
                zoom: 'Zoom',
                back: 'Step back'
              }
            },
            restore: {
              show: true,
              title: 'Restore',
            }
          }
        },
        dataZoom: [
          {
            type: 'inside',
            start: 0,
            end: 100
          },
          {
            show: true,
            type: 'slider',
            start: 0,
            end: 100,
            height: 30,
            bottom: 0,
            borderColor: '#ccc',
            fillerColor: 'rgba(40,54,139,0.21)',
            handleStyle: {
              color: '#004b8b'
            }
          }
        ],
        calculable: true,
        yAxis: [{
          type: 'value',
          boundaryGap: [0, 0.01],
          itemStyle: {
            normal: {
              label: {
                rotate: 90
              }
            }
          }
        }],
        xAxis: [
          {
            type: 'category',
            data: option.legend
          }
        ],
        series: [
          {
            name: 'Quantity',
            type: this.chartType,
            barMaxWidth: '25px',
            data: option.values,
          }
        ]
      };
      this.loadingOption = false;
    });
  }

  private onSuccess(data, headers) {
    this.data = data;
    this.buildCharTopFieldValue(this.data);
  }

  resolveFieldName(): string {
    return (this.field.type === ElasticDataTypesEnum.TEXT ||
      this.field.type === ElasticDataTypesEnum.STRING) ? this.field.name + '.keyword' : this.field.name;
  }

  private onError(status, header) {
    if (status === 500 && header.get('x-utmvaultApp-code') === '1000') {
      this.utmToastService.showWarning('This interval create too many buckets to show in the' +
        ' selected time range, so it has been scaled to default', 'Too many buckets');
      this.interval = 'DAY';
      this.getFieldTopValue();
    }
  }

  private processTopFieldValue(data: LogAnalyzerChartResponseType): Observable<ChartModel> {
    return new Observable<ChartModel>((resolver) => {
      const config: ChartModel = {
        series: [], values: [], legend: []
      };
      if (data && data.values.length > 0) {
        config.legend = data.categories;
        config.values = data.values;
        resolver.next(config);
      } else {
        this.loadingOption = false;
        resolver.next(null);
      }
    });
  }

}
