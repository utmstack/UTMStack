import {Component, Input, OnInit} from '@angular/core';
import {NetScanSoftwares} from '../../types/net-scan-softwares';

@Component({
  selector: 'app-asset-softwares',
  templateUrl: './asset-softwares.component.html',
  styleUrls: ['./asset-softwares.component.scss']
})
export class AssetSoftwaresComponent implements OnInit {
  @Input() softwares: NetScanSoftwares[];
  totalItems: number;
  page = 1;
  itemsPerPage = 10;
  pageStart = 0;
  pageEnd = 10;

  constructor() {
  }

  ngOnInit() {
    this.softwares = this.softwares ? this.softwares : [];
    this.totalItems = this.softwares ? this.softwares.length : 0;
  }

  loadPage(page: number) {
    this.pageEnd = this.page * this.itemsPerPage;
    this.pageStart = this.pageEnd - this.itemsPerPage;
  }
}


