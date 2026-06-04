import {Component, OnDestroy, OnInit} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {JhiMetricsService} from './metrics.service';

@Component({
  selector: 'app-app-metrics',
  templateUrl: './app-metrics.component.html',
  styleUrls: ['./app-metrics.component.scss']
})
export class AppMetricsComponent implements OnInit, OnDestroy {
  metrics: any = {};
  threadData: any;
  updatingMetrics = true;
  JCACHE_KEY: string;
  interval: any;

  constructor(private modalService: NgbModal, private metricsService: JhiMetricsService) {
    this.JCACHE_KEY = 'jcache.statistics';
  }

  ngOnInit() {
    this.refresh();
    this.interval = setInterval(() => this.refresh(), 10000);
  }

  ngOnDestroy() {
    clearInterval(this.interval);
  }

  refresh() {
    this.metricsService.getMetrics().subscribe(metrics => {
      this.metrics = metrics;
      this.updatingMetrics = false;
    });
  }

  isObjectExisting(metrics: any, key: string) {
    return metrics && metrics[key];
  }

  get httpRequest() {
    return this.metrics['http.server.requests'];
  }

  get services() {
    return this.metrics.services;
  }

  isObjectExistingAndNotEmpty(metrics: any, key: string) {
    return this.isObjectExisting(metrics, key) && JSON.stringify(metrics[key]) !== '{}';
  }
}
