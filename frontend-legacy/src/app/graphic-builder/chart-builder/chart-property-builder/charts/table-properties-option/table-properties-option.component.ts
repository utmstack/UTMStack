import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup} from '@angular/forms';
import {UtmTableOptionType} from '../../../../../shared/chart/types/charts/table/utm-table-option.type';
import {VisualizationType} from '../../../../../shared/chart/types/visualization.type';
import {ChartTypeEnum} from '../../../../../shared/enums/chart-type.enum';

@Component({
  selector: 'app-table-properties-option',
  templateUrl: './table-properties-option.component.html',
  styleUrls: ['./table-properties-option.component.scss']
})
export class TablePropertiesOptionComponent implements OnInit {
  @Input() visualization: VisualizationType;
  @Input() mode: string;
  @Output() tableOptions = new EventEmitter<UtmTableOptionType>();
  tableForm: FormGroup;
  aggs = ['sum', 'avg', 'max', 'min', 'count'];
  charTypeEnum = ChartTypeEnum;

  constructor(private fb: FormBuilder) {
  }

  ngOnInit() {
    this.initFormTable();
    if (this.mode === 'edit') {
      if (typeof this.visualization.chartConfig === 'string') {
        this.visualization.chartConfig = JSON.parse(this.visualization.chartConfig);
      }
      this.tableForm.patchValue(this.visualization.chartConfig, {emitEvent: true});
      if (!this.visualization.chartConfig.hasOwnProperty('dynamicPageSize')) {
        this.tableForm.get('dynamicPageSize').setValue(false);
      }
    }
    this.tableForm.valueChanges.subscribe(value => {
      this.tableOptions.emit(value);
    });
    this.tableOptions.emit(this.tableForm.value);
  }

  initFormTable() {
    this.tableForm = this.fb.group(
      {
        itemsPerPage: [10],
        showTotal: [false],
        totalFunction: ['sum'],
        exportCsv: [true],
        dynamicPageSize: [true]
      }
    );
  }
}
