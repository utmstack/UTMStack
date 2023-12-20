import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup} from '@angular/forms';
import {DataZoom} from '../../../../../shared/chart/types/charts/chart-properties/datazoom/data-zoom';
import {DATA_ZOOM_ORIENT} from '../../../../shared/const/chart/datazoom-properties.const';

@Component({
  selector: 'app-chart-data-zoom',
  templateUrl: './chart-data-zoom.component.html',
  styleUrls: ['./chart-data-zoom.component.scss']
})
export class ChartDataZoomComponent implements OnInit {
  @Output() dataZoomOption = new EventEmitter<DataZoom>();
  orientation = DATA_ZOOM_ORIENT;
  formDataZoom: FormGroup;
  viewDataZoom = false;

  constructor(private fb: FormBuilder) {
  }

  ngOnInit() {
    this.initFormDataZom();
    this.formDataZoom.valueChanges.subscribe(value => {
      this.dataZoomOption.emit(this.formDataZoom.value);
    });
    this.dataZoomOption.emit(this.formDataZoom.value);
  }

  initFormDataZom() {
    this.formDataZoom = this.fb.group({
      show: [false],
      orient: ['horizontal'],
      type: ['slider'],
      start: [0],
      end: [100],
      height: [40],
      width: [],
      borderColor: ['#ccc'],
      fillerColor: ['rgba(40,54,139,0.21)'],
      handleStyle: this.fb.group({
        color: ['#004b8b']
      }),
      left: [],
      top: [],
      right: [],
      bottom: [],
    });
  }


  viewDataZoomProperties() {
    this.viewDataZoom = this.viewDataZoom ? false : true;
  }
}
