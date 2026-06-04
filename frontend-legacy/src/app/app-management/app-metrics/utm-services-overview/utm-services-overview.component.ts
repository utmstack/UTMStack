import {Component, Input, OnChanges, OnInit, SimpleChanges} from '@angular/core';


@Component({
  selector: 'app-utm-services-overview',
  templateUrl: './utm-services-overview.component.html',
  styleUrls: ['./utm-services-overview.component.scss']
})
export class UtmServicesOverviewComponent implements OnInit, OnChanges {
  @Input() serviceData: Map<string, Map<string, { max: number, mean: number, count: number }>>;
  services: { path: string, method: string, mean: number, count: number }[] = [];
  page = 1;
  itemsPerPage = 10;
  pageStart = 0;
  pageEnd = 10;
  totalItems: number;

  constructor() {
  }

  ngOnInit() {
    this.convertServiceDataToArray(this.serviceData);
  }

  ngOnChanges(changes: SimpleChanges) {
    if (changes && changes.serviceData && !changes.serviceData.firstChange) {
      this.convertServiceDataToArray(changes.serviceData.currentValue);
    }
  }

  loadPage(page: number) {
    this.pageEnd = page * this.itemsPerPage;
    this.pageStart = this.pageEnd - this.itemsPerPage;
  }

  convertServiceDataToArray(serviceData: Map<string, Map<string,
    { max: number, mean: number, count: number }>>) {
    const result: { path: string, method: string, mean: number, count: number }[] = [];
    Object.keys(serviceData).forEach(path => {
      Object.keys(serviceData[path]).forEach(method => {
        const {mean, count} = serviceData[path][method];
        result.push({path, method, mean, count});
      });
    });
    this.services = result.filter(value => value.path !== '/**' && value.path !== 'root')
      .sort((a, b) => a.count < b.count ? 1 : -1);
    this.totalItems = this.services ? this.services.length : 0;
  }
}
