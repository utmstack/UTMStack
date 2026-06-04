import {Component, Input, OnInit} from '@angular/core';

@Component({
  selector: 'app-utm-progressbar',
  templateUrl: './utm-progressbar.component.html',
  styleUrls: ['./utm-progressbar.component.css']
})
export class UtmProgressbarComponent implements OnInit {
  @Input() value: number;
  @Input() threshold = {
    low: 50,
    medium: 75,
    high: 85
  };

  constructor() {
  }

  ngOnInit() {
  }

  getColorLine(): string {
    if (this.value <= this.threshold.low) {
      return 'success';
    } else if (this.value > this.threshold.low && this.value <= this.threshold.medium) {
      return 'primary';
    } else if (this.value > this.threshold.medium && this.value <= this.threshold.high) {
      return 'warning';
    } else if (this.value > this.threshold.high) {
      return 'danger';
    } else {
      return 'dark';
    }
  }
}
