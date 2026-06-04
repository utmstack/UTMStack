import {Component, Input, OnInit} from '@angular/core';

@Component({
  selector: 'app-utm-jvm-metrics',
  templateUrl: './utm-jvm-metrics.component.html',
  styleUrls: ['./utm-jvm-metrics.component.scss']
})
export class UtmJvmMetricsComponent implements OnInit {
  @Input() jvm: object;
  @Input() loading: boolean;

  constructor() {
  }

  ngOnInit() {
  }

}
