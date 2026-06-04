import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';

@Component({
  selector: 'app-gauge-range-color',
  templateUrl: './gauge-range-color.component.html',
  styleUrls: ['./gauge-range-color.component.scss']
})
export class GaugeRangeColorComponent implements OnInit {
  @Output() rangeColorChange = new EventEmitter<[number, string][]>();
  @Input() ranges: [number, string][];
  viewGaugeColor = false;

  constructor() {
  }

  ngOnInit() {
    this.rangeColorChange.emit(this.ranges);
  }

  viewGaugeColorRangeProperties() {
    this.viewGaugeColor = this.viewGaugeColor ? false : true;
  }

  deleteRange(index: number) {
    this.ranges.splice(index, 1);
    this.rangeColorChange.emit(this.ranges);
  }

  addRange() {
    this.ranges.push([0.9, '#555555']);
    this.sortRanges();
  }

  onRangeChange($event, index: number) {
    this.ranges[index][0] = this.ranges[index][0] = Number($event.target.value);
    this.sortRanges();
  }

  sortRanges() {
    this.ranges.sort((a, b) => {
      return a[0] - b[0];
    });
    this.rangeColorChange.emit(this.ranges);
  }

}

