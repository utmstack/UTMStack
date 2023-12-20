import {Component, OnDestroy, OnInit} from '@angular/core';
import {AccountService} from '../../../core/auth/account.service';
import {ElasticHealthService} from '../../services/elasticsearch/elastic-health.service';
import {ElasticHealthStatsType} from '../../types/elasticsearch/elastic-health-stats.type';

@Component({
  selector: 'app-utm-header-health-warning',
  templateUrl: './utm-header-health-warning.component.html',
  styleUrls: ['./utm-header-health-warning.component.scss']
})
export class UtmHeaderHealthWarningComponent implements OnInit, OnDestroy {
  clusterHealth: ElasticHealthStatsType;
  interval: any;
  show = true;

  constructor(private elasticHealthService: ElasticHealthService, private accountService: AccountService) {
  }

  ngOnInit() {
    this.accountService.identity().then(account => {
      if (account) {
        this.initInterval();
      }
    });
  }

  ngOnDestroy() {
    clearInterval(this.interval);
  }

  getHealth() {
    this.elasticHealthService.queryHealth().subscribe(response => {
      this.clusterHealth = response.body.resume;
    });
  }

  initInterval() {
    this.interval = setInterval(() => {
      this.accountService.identity().then(account => {
        if (account) {
          this.getHealth();
        }
      });
    }, 10000);
  }

  hide() {
    clearInterval(this.interval);
    this.show = false;
    setTimeout(() => {
      this.initInterval();
    }, 36000);
  }
}
