import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormArray, FormBuilder, FormGroup} from '@angular/forms';
import {Observable} from 'rxjs';
import {SeriesLine} from '../../../../../shared/chart/types/charts/chart-properties/series/line/series-line';
import {MetricAggregationType} from '../../../../../shared/chart/types/metric/metric-aggregation.type';
import {ChartTypeEnum} from '../../../../../shared/enums/chart-type.enum';
import {MetricDataBehavior} from '../../shared/behaviors/metric-data.behavior';
import {extractLabelFromMetric} from '../../shared/functions/visualization-util';

@Component({
  selector: 'app-chart-series-line-bar-option',
  templateUrl: './chart-series-line-bar-option.component.html',
  styleUrls: ['./chart-series-line-bar-option.component.scss']
})
export class ChartSeriesLineBarOptionComponent implements OnInit {
  @Output() seriesOptions = new EventEmitter<SeriesLine>();
  @Input() chartType: ChartTypeEnum;
  chartTypeEnum = ChartTypeEnum;
  formSerie: FormGroup;
  viewSerie = false;
  types = ['line', 'bar'];
  metrics: MetricAggregationType[];

  constructor(private fb: FormBuilder,
              private metricDataBehavior: MetricDataBehavior) {
  }

  get series() {
    return this.formSerie.controls.series as FormArray;
  }

  ngOnInit() {
    this.initFormLine();
    this.metricDataBehavior.$metricDeletedId.subscribe(id => {
      if (id !== -1) {
        const indexOption = this.series.controls.findIndex(value =>
          value.get('metricId').value === id);
        if (indexOption > -1) {
          this.deleteSerie(indexOption);
        }
      }
    });
    this.metricDataBehavior.$metric.subscribe(met => {
      if (met.length > 0) {
        this.metrics = met;
        this.addSeries(met).subscribe(value => {
          this.seriesOptions.emit(value);
        });
      }
    });

    this.formSerie.valueChanges.subscribe(value => {
      this.seriesOptions.emit(this.formSerie.get('series').value);
    });
  }

  initFormLine() {
    this.formSerie = this.fb.group({
      series: this.fb.array([]),
    });
  }

  deleteSerie(index: number) {
    this.series.removeAt(index);
  }

  viewSerieProperties() {
    this.viewSerie = this.viewSerie ? false : true;
  }

  getSerieLabel(index: number): string {
    const i = this.series.at(index).get('metricId').value;
    return extractLabelFromMetric(i, this.metrics);
  }

  private addSeries(metrics: MetricAggregationType[]): Observable<SeriesLine> {
    return new Observable<SeriesLine>(subscriber => {
      for (const metric of metrics) {
        const indexOption = this.series.controls.findIndex(value =>
          value.get('metricId').value === metric.id);
        if (indexOption === -1) {
          this.series.push(this.fb.group({
            metricId: [metric.id],
            name: '',
            type: [(this.chartType === this.chartTypeEnum.AREA_LINE_CHART ||
              this.chartType === this.chartTypeEnum.LINE_CHART) ? 'line' : 'bar'],
            data: [[]],
            smooth: true,
            markPoint: [null],
            markLine: [null],
            itemStyle: [
              this.chartType === this.chartTypeEnum.AREA_LINE_CHART ? {
                normal: {
                  areaStyle: {
                    type: 'mint',
                    normal: {
                      opacity: 0.25
                    }
                  }
                }
              } : null]
          }));
        }
      }
      subscriber.next(this.formSerie.get('series').value);
    });

  }

  private setChartType() {
    switch (this.chartType) {

    }
  }
}
