import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormArray, FormBuilder, FormGroup} from '@angular/forms';
import {SeriesMark} from '../../../../../shared/chart/types/charts/chart-properties/series/series-properties/marks/series-mark';

@Component({
  selector: 'app-chart-mark-option',
  templateUrl: './chart-mark-option.component.html',
  styleUrls: ['./chart-mark-option.component.scss']
})
export class ChartMarkOptionComponent implements OnInit {
  @Input() markType: 'point' | 'line';
  @Output() markChange = new EventEmitter<SeriesMark>();
  markForm: FormGroup;
  symbols = ['circle', 'rect', 'roundRect', 'triangle', 'diamond', 'pin', 'arrow', 'none'];
  viewMark = false;
  pointTypes = ['min', 'max', 'average'];

  constructor(private fb: FormBuilder) {
  }

  get data() {
    return this.markForm.controls.data as FormArray;
  }

  ngOnInit() {
    this.initMarkForm();
    this.markForm.valueChanges.subscribe(value => {
      this.markChange.emit(value);
    });
    if (this.markType === 'point') {
      this.markForm.get('symbol').setValue('pin');
      this.markForm.get('symbolSize').setValue(35);
    }
  }

  initMarkForm() {
    this.markForm = this.fb.group({
      symbol: [null],
      symbolSize: [null],
      label: this.fb.group({
        fontSize: [10]
      }),
      data: this.fb.array([]),
    });
  }

  delete(index: number) {
    this.data.removeAt(index);
  }

  addData() {
    this.data.push(this.fb.group({
      type: ['min'],
      name: ['']
    }));
  }

  viewMarkProperties() {
    this.viewMark = this.viewMark ? false : true;
  }
}
