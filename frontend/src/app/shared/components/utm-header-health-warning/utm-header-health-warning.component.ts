import {ChangeDetectionStrategy, Component, OnInit} from '@angular/core';
import {Observable} from 'rxjs';
import {filter, map, tap} from 'rxjs/operators';
import {AccountService} from '../../../core/auth/account.service';
import {ElasticHealthService} from '../../services/elasticsearch/elastic-health.service';
import {ElasticHealthStatsType} from '../../types/elasticsearch/elastic-health-stats.type';

@Component({
  selector: 'app-utm-header-health-warning',
  templateUrl: './utm-header-health-warning.component.html',
  styleUrls: ['./utm-header-health-warning.component.scss'],
  changeDetection: ChangeDetectionStrategy.Default
})
export class UtmHeaderHealthWarningComponent implements OnInit {
  clusterHealth$: Observable<ElasticHealthStatsType>;
  interval: any;
  show = true;

  constructor(private elasticHealthService: ElasticHealthService, private accountService: AccountService) {
  }

  ngOnInit() {
    this.accountService.identity().then(account => {
      if (account) {
        this.getHealth();
      }
    });
  }

  getHealth() {
    this.clusterHealth$ = this.elasticHealthService.health$
      .pipe(
        filter(response => !!response),
        tap( () => this.show = true),
        map (response => {
          return {...response.resume};
        }));
  }

  hide() {
    this.show = false;
  }
}
