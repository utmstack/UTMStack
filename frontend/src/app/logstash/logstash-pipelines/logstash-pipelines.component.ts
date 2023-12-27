import {AfterViewChecked, ChangeDetectorRef, Component, OnDestroy, OnInit} from '@angular/core';
import {ElasticHealthService} from '../../shared/services/elasticsearch/elastic-health.service';
import {LogstashService} from '../../shared/services/logstash/logstash.service';
import {ElasticHealthStatsType} from '../../shared/types/elasticsearch/elastic-health-stats.type';
import {getElasticClusterHealth} from '../../shared/util/elastic-health.util';
import {SVG_MODULE_MAP} from '../shared/const/logstash.const';
import {UtmLogstashPipeline, UtmPipeline} from '../shared/types/logstash-stats.type';

@Component({
  selector: 'app-logstash-pipelines',
  templateUrl: './logstash-pipelines.component.html',
  styleUrls: ['./logstash-pipelines.component.scss']
})
export class LogstashPipelinesComponent implements OnInit, OnDestroy, AfterViewChecked {
  logstashPipelines: UtmLogstashPipeline;
  pipelineDetail: UtmPipeline = null;
  clusterHealth: ElasticHealthStatsType;
  svgMap = SVG_MODULE_MAP;
  interval: any;
  imageChangeInterval: any;
  images: string[] = [];
  img1: string;
  img2: string;
  img3: string;
  img4: string;
  img5: string;

  constructor(private elasticHealthService: ElasticHealthService,
              private logstashService: LogstashService,
              private cdr: ChangeDetectorRef) {
  }

  ngOnInit() {
    this.getHealth();
    this.getLogstashStats();
    this.interval = setInterval(() => {
      this.getHealth();
      this.getLogstashStats();
    }, 5000);
    this.imageChangeInterval = setInterval(() => this.setImages(), 8000);
  }

  ngAfterViewChecked() {
    this.cdr.detectChanges();
  }

  ngOnDestroy() {
    clearInterval(this.interval);
    clearInterval(this.imageChangeInterval);
  }

  getHealth() {
    this.elasticHealthService.queryHealth().subscribe(response => {
      this.clusterHealth = response.body.resume;
    });
  }

  getLogstashStats() {
    this.logstashService.getStats().subscribe(response => {
      this.logstashPipelines = response.body;
      this.images = this.logstashPipelines.pipelines.map(value => value.moduleName);
      if (!this.img1) {
        this.setImages();
      }
    });
  }

  viewPipeline(pipeline: UtmPipeline) {
    this.pipelineDetail = pipeline;
  }

  closeDetail() {
    this.pipelineDetail = null;
  }

  getClusterHealth() {
    return getElasticClusterHealth(this.clusterHealth, 'UP');
  }

  get getRandomSvg(): string {
    const randomModule = this.images[Math.floor(Math.random() * this.images.length)];
    return this.getModuleSvg(randomModule);
  }

  setImages() {
    const randomImage1 = this.images[Math.floor(Math.random() * this.images.length)];
    const randomImage2 = this.images[Math.floor(Math.random() * this.images.length)];
    const randomImage3 = this.images[Math.floor(Math.random() * this.images.length)];
    const randomImage4 = this.images[Math.floor(Math.random() * this.images.length)];
    const randomImage5 = this.images[Math.floor(Math.random() * this.images.length)];
    this.img1 = this.getModuleSvg(randomImage1);
    this.img2 = this.getModuleSvg(randomImage2);
    this.img3 = this.getModuleSvg(randomImage3);
    this.img4 = this.getModuleSvg(randomImage4);
    this.img5 = this.getModuleSvg(randomImage5);
    this.cdr.detectChanges();
  }

  getModuleSvg(module: string): string {
    const icon = this.svgMap[module];
    return icon ? icon : 'generic.svg';
  }
}
