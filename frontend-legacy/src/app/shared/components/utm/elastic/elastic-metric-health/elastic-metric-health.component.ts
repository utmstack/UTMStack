import {Component, Input, OnInit} from '@angular/core';
import {ElasticHealthStatsType} from '../../../../types/elasticsearch/elastic-health-stats.type';

@Component({
  selector: 'app-elastic-metric-health',
  templateUrl: './elastic-metric-health.component.html',
  styleUrls: ['./elastic-metric-health.component.css']
})
export class ElasticMetricHealthComponent implements OnInit {
  @Input() clusterHealth: ElasticHealthStatsType;
  @Input() status: string;

  constructor() {
  }

  ngOnInit() {
  }

}
