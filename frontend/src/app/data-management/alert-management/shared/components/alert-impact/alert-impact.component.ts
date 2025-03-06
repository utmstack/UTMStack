import {Component, Input, OnInit} from '@angular/core';

@Component({
  selector: 'app-alert-impact',
  templateUrl: './alert-impact.component.html',
  styleUrls: ['./alert-impact.component.scss']
})
export class AlertImpactComponent implements OnInit {
  @Input() severity: string | number;
  label: string;
  background: string;

  constructor() {
  }

  ngOnInit() {

  }
}
