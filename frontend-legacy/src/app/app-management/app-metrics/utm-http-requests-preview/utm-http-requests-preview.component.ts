import {Component, Input, OnInit} from '@angular/core';

@Component({
  selector: 'app-utm-http-requests-preview',
  templateUrl: './utm-http-requests-preview.component.html',
  styleUrls: ['./utm-http-requests-preview.component.css']
})
export class UtmHttpRequestsPreviewComponent implements OnInit {
  @Input() httpPreview: {
    all: { count: number }
    percode: Map<number, { max: number, mean: number, count: number }>
  };

  constructor() {
  }

  ngOnInit() {
  }

}
