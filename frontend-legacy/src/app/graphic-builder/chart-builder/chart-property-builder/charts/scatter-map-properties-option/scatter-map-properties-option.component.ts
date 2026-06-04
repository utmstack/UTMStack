import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup} from '@angular/forms';
import {Tooltip} from '../../../../../shared/chart/types/charts/chart-properties/tooltip/tooltip';
import {VisualMap} from '../../../../../shared/chart/types/charts/chart-properties/visualmap/visual-map';
import {UtmScatterMapOptionType} from '../../../../../shared/chart/types/charts/scatter/utm-scatter-map-option.type';

@Component({
  selector: 'app-scatter-map-properties-option',
  templateUrl: './scatter-map-properties-option.component.html',
  styleUrls: ['./scatter-map-properties-option.component.scss']
})
export class ScatterMapPropertiesOptionComponent implements OnInit {
  @Output() scatterMapOptions = new EventEmitter<UtmScatterMapOptionType>();
  formScatterOption: FormGroup;
  viewScatterOptions = false;

  constructor(private fb: FormBuilder) {
  }

  ngOnInit() {
    this.initFormScatter();
    this.formScatterOption.valueChanges.subscribe(value => {
      this.scatterMapOptions.emit(value);
    });

    this.scatterMapOptions.emit(this.formScatterOption.value);
  }

  initFormScatter() {
    this.formScatterOption = this.fb.group({
      tooltip: [new Tooltip('item')],
      visualMap: new VisualMap(),
      leaflet: [],
      series: [],
      top: [10]
    });
  }

  viewScatterOptionProperties() {
    this.viewScatterOptions = this.viewScatterOptions ? false : true;
  }
}
