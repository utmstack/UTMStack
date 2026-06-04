import {Component, Input, OnInit} from '@angular/core';

@Component({
  selector: 'app-utm-table-detail',
  templateUrl: './utm-table-detail.component.html',
  styleUrls: ['./utm-table-detail.component.scss']
})
export class UtmTableDetailComponent implements OnInit {
  @Input() row: any;
  viewMode = 'table';
  private detailWidth: number;

  constructor() {
    this.detailWidth = window.innerWidth - 300;
  }

  ngOnInit() {
  }

}
