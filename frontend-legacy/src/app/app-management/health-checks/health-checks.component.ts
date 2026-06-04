import {Component, OnInit} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {ElasticHealthService} from '../../shared/services/elasticsearch/elastic-health.service';
import {ElasticHealthStatsType} from '../../shared/types/elasticsearch/elastic-health-stats.type';
import {UtmHealthService} from './health-checks.service';
import {HealthDetailComponent} from './health-detail/health-detail.component';

@Component({
  selector: 'app-health-checks',
  templateUrl: './health-checks.component.html',
  styleUrls: ['./health-checks.component.scss']
})
export class HealthChecksComponent implements OnInit {
  healthData: any;
  updatingHealth: boolean;
  clusterNodes: ElasticHealthStatsType[];

  constructor(private modalService: NgbModal,
              private elasticHealthService: ElasticHealthService,
              private healthService: UtmHealthService) {
  }

  ngOnInit() {
    this.refresh();
  }

  getHealth() {
    this.elasticHealthService.queryHealth().subscribe(response => {
      this.clusterNodes = response.body.nodes;
    });
  }

  baseName(name: string) {
    return this.healthService.getBaseName(name);
  }

  getBadgeClass(statusState) {
    if (statusState === 'UP') {
      return 'badge-success';
    } else {
      return 'badge-danger';
    }
  }

  refresh() {
    this.updatingHealth = true;
    this.getHealth();
    this.healthService.checkHealth().subscribe(
      health => {
        delete (health.components.elasticsearch);
        this.healthData = this.healthService.transformHealthData(health.components);
        this.updatingHealth = false;
      },
      error => {
        if (error.status === 503) {
          this.healthData = this.healthService.transformHealthData(error.error);
          this.updatingHealth = false;
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
}
