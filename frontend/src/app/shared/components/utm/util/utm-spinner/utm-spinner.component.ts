import {Component, Input, OnInit} from '@angular/core';

@Component({
    standalone: false,
    selector: 'app-utm-spinner',
    templateUrl: './utm-spinner.component.html',
    styleUrls: ['./utm-spinner.component.scss'],
})
export class UtmSpinnerComponent implements OnInit {
  @Input() width;
  @Input() height;
  @Input() loading;
  @Input() label: string;
  @Input() color: string;
  @Input() margin: string;

  constructor() {
  }

  ngOnInit() {
  }

}
