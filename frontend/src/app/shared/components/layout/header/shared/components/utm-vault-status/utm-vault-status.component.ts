import {Component, Input, OnDestroy, OnInit} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {NgxSpinnerService} from 'ngx-spinner';
import {UtmHealthService} from '../../../../../../../app-management/health-checks/health-checks.service';
import {HealthDetailComponent} from '../../../../../../../app-management/health-checks/health-detail/health-detail.component';
import {AccountService} from '../../../../../../../core/auth/account.service';
import {UtmLogstashPipeline} from '../../../../../../../logstash/shared/types/logstash-stats.type';
import {UtmToastService} from '../../../../../../alert/utm-toast.service';
import {ElasticHealthService} from '../../../../../../services/elasticsearch/elastic-health.service';
import {LogstashService} from '../../../../../../services/logstash/logstash.service';
import {ElasticHealthStatsType} from '../../../../../../types/elasticsearch/elastic-health-stats.type';
import {UtmHealthType} from '../../../../../../types/utm-health.type';
import {getElasticClusterHealth} from '../../../../../../util/elastic-health.util';

@Component({
  selector: 'app-utm-vault-status',
  templateUrl: './utm-vault-status.component.html',
  styleUrls: ['./utm-vault-status.component.scss']
})
export class UtmVaultStatusComponent implements OnInit, OnDestroy {
  @Input() version: 'mobile' | 'desktop';
  healthData: any[];
  updatingHealth = true;
  interval: any;
  errorCount = 0;
  clusterHealth: ElasticHealthStatsType;
  health: UtmHealthType;
  logstashPipelines: UtmLogstashPipeline;

  constructor(private modalService: NgbModal,
              private toastService: UtmToastService,
              private spinner: NgxSpinnerService,
              private accountService: AccountService,
              private elasticHealthService: ElasticHealthService,
              private logstashService: LogstashService,
              private healthService: UtmHealthService) {
  }

  ngOnInit() {
    this.refresh();
    this.getHealth();
    this.getLogstashStats();
    this.interval = setInterval(() => {
      this.accountService.identity().then(account => {
        if (account) {
          this.refresh();
          this.getHealth();
          this.getLogstashStats();
        }
      });
    }, 10000);
  }

  ngOnDestroy(): void {
    clearInterval(this.interval);
  }

  baseName(name: string) {
    return this.healthService.getBaseName(name);
  }

  getLogstashStats() {
    this.logstashService.getStats().subscribe(response => {
      this.logstashPipelines = response.body;
    });
  }

  getBadgeClass(statusState) {
    if (statusState === 'UP') {
      return 'badge-success';
    } else {
      return 'badge-danger';
    }
  }

  refresh() {
    this.healthService.checkHealth().subscribe(
      response => {
        this.health = response;
        this.updatingHealth = false;
        this.getAppHealth();
      },
      error => {
        if (error.status === 503) {
          // this.healthData = this.healthService.transformHealthData(error.error);
          this.updatingHealth = false;
          // this.getAppHealth(this.healthData);
        }
      }
    );
  }

  showHealth(health: any) {
    const modalRef = this.modalService.open(HealthDetailComponent, {centered: true, size: 'sm'});
    modalRef.componentInstance.currentHealth = health;
    modalRef.result.then(
      result => {
        // Left blank intentionally, nothing to do here
      },
      reason => {
        // Left blank intentionally, nothing to do here
      }
    );
  }

  subSystemName(name: string) {
    return this.healthService.getSubSystemName(name);
  }

  appStatus(): boolean {
    return this.resolveAppStatus();
  }

  resolveAppStatus(): boolean {
    let status = true;
    this.healthData.forEach(value => {
      if (value.status !== 'UP') {
        status = false;
      }
    });
    return status;
  }

  resolveIcon(health: any) {
    switch (health) {
      case 'diskSpace':
        return '/assets/icons/header/database_hardisk.svg';
      case 'elasticsearch':
        return '/assets/icons/header/database_performance.svg';
      case 'db':
        return '/assets/icons/header/database_connection.svg';
    }
  }

  getAppHealth() {
    for (const key of Object.keys(this.health.components)) {
      if (this.health.components[key].status !== 'UP') {
        this.notify(this.health.components[key].status, key);
      }
    }
  }

  notify(status, key) {
    this.toastService.showError('Notice', 'Service ' +
      this.resolveServiceName(key) + ' is ' + status
      + ' we are trying to reconnect with this service, in case the problem persists contact with the support team');
  }

  isElasticHealth(health) {
    return this.baseName(health.name).toLowerCase().includes('engine');
  }

  resolveServiceName(key: any) {
    switch (key) {
      case 'diskSpace':
        return 'Disk Space';
      case 'elasticsearch':
        return 'Data Engine';
      case 'db':
        return 'Database';
    }
  }

  getHealth() {
    this.elasticHealthService.queryHealth().subscribe(response => {
      this.clusterHealth = response.body.resume;
    });
  }

  getClusterHealth(): 'CRITIC' | 'MEDIUM' | 'UP' | 'DOWN' {
    const connectionStatus = this.getElasticConnectionStatus();
    return getElasticClusterHealth(this.clusterHealth, connectionStatus);
  }

  getElasticConnectionStatus(): 'UP' | 'DOWN' {
    return this.health.components.elasticsearch.status;
  }

  convertToGb(bytes: number): string {
    const val = bytes / 1073741824;
    if (val > 1) {
      // Value
      return val.toFixed(2) + ' GB';
    } else {
      return (bytes / 1048576).toFixed(2) + ' MB';
    }
  }


  calcPercentDiskSpace() {
    const used = this.calcDiskUsed();
    return used / this.health.components.diskSpace.details.total * 100;
  }

  calcDiskUsed(): number {
    return this.health.components.diskSpace.details.total - this.health.components.diskSpace.details.free;
  }
}
