import {Component, Input, OnInit} from '@angular/core';
import {NetScanPortsType} from '../../types/net-scan-ports.type';

@Component({
  selector: 'app-assets-ports',
  templateUrl: './assets-ports.component.html',
  styleUrls: ['./assets-ports.component.scss']
})
export class AssetsPortsComponent implements OnInit {
  @Input() ports: NetScanPortsType[];
  totalItems: number;
  page = 1;
  itemsPerPage = 10;
  pageStart = 0;
  pageEnd = 10;

  constructor() {
  }

  ngOnInit() {
    this.totalItems = this.ports.length;
  }

  loadPage(page: number) {
    this.pageEnd = this.page * this.itemsPerPage;
    this.pageStart = this.pageEnd - this.itemsPerPage;
  }
}
