import {Component, OnDestroy, OnInit} from '@angular/core';

@Component({
  selector: 'app-active-directory-dashboard',
  templateUrl: './active-directory-dashboard.component.html',
  styleUrls: ['./active-directory-dashboard.component.scss']
})
export class AdDashboardComponent implements OnInit, OnDestroy {
  pdfExport = false;
  refreshInterval = 60000;

  constructor() {
  }

  ngOnInit() {
    window.addEventListener('beforeprint', (event) => {
      this.pdfExport = true;
    });
    window.addEventListener('afterprint', (event) => {
      this.pdfExport = false;
    });
  }

  ngOnDestroy(): void {
  }

  exportToPdf() {
    this.pdfExport = true;
    // captureScreen('utmDashboardActiveDirectory').then((finish) => {
    //   this.pdfExport = false;
    // });
    setTimeout(() => {
      window.print();
    }, 1000);
  }
}


export interface ElasticPieResponse {
  metricId: string;
  bucketKey: string;
  value: number;
  bucketId: string;
}
