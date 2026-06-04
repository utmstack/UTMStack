import {Component, OnInit} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {ElasticHealthService} from '../../../shared/services/elasticsearch/elastic-health.service';
import {ElasticHealthType} from '../../../shared/types/elasticsearch/elastic-health.type';

@Component({
  selector: 'app-health-cluster',
  templateUrl: './health-cluster.component.html',
  styleUrls: ['./health-cluster.component.css']
})
export class HealthClusterComponent implements OnInit {
  clusterHealth: ElasticHealthType;

  constructor(private elasticHealthService: ElasticHealthService,
              public activeModal: NgbActiveModal) {
  }


  ngOnInit() {
    this.getHealth();
  }

  getHealth() {
    this.elasticHealthService.queryHealth().subscribe(response => {
      this.clusterHealth = response.body;
    });
  }

}
