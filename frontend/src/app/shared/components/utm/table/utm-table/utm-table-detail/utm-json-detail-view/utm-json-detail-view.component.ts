import {Component, Input, OnInit} from '@angular/core';

@Component({
  selector: 'app-utm-json-detail-view',
  templateUrl: './utm-json-detail-view.component.html',
  styleUrls: ['./utm-json-detail-view.component.scss']
})
export class UtmJsonDetailViewComponent implements OnInit {
  @Input() rowDocument: any;
  detailWidth: number;

  constructor() {
    this.detailWidth = window.innerWidth - 330;
  }

  ngOnInit() {
  }

}
