import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';

@Component({
  selector: 'app-kpi',
  templateUrl: './kpi.component.html',
  styleUrls: ['./kpi.component.scss']
})
export class KpiComponent implements OnInit {
  @Input() label: string;
  @Input() group: string;
  @Input() value: number;
  @Input() icon: string;
  @Input() color: string;
  @Input() prevValue: number;
  @Input() decimal: number;
  @Output() metricClick = new EventEmitter<string>();

  constructor() {
  }

  ngOnInit() {
    if (this.decimal) {
      this.value.toPrecision(this.decimal);
    }
  }

  emitValue() {
    this.metricClick.emit(this.label);
  }
}
