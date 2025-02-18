import {Component, Input, OnInit} from '@angular/core';
import {UtmLogstashPipeline} from '../../../../../logstash/shared/types/logstash-stats.type';
import {convertBytesToGB} from '../../../../util/bytes.util';

@Component({
  selector: 'app-logstash-stats',
  templateUrl: './logstash-stats.component.html',
  styleUrls: ['./logstash-stats.component.scss']
})
export class LogstashStatsComponent implements OnInit {
  @Input() logstashPipelines: UtmLogstashPipeline;

  constructor() {
  }

  ngOnInit() {
  }

  get memoryPercentage(): number {
    if (
      !this.logstashPipelines ||
      !this.logstashPipelines.general ||
      !this.logstashPipelines.general.jvm ||
      !this.logstashPipelines.general.jvm.mem
    ) {
      return 0;
    }

    const part = this.logstashPipelines.general.jvm.mem.nonHeapUsedInBytes;
    const total = this.logstashPipelines.general.jvm.mem.nonHeapCommittedInBytes || 1; // Evita división por 0
    return parseFloat(((part / total) * 100).toFixed(2));
  }

  get memoryUsed(): number {
    if (
      !this.logstashPipelines ||
      !this.logstashPipelines.general ||
      !this.logstashPipelines.general.jvm ||
      !this.logstashPipelines.general.jvm.mem
    ) {
      return 0;
    }

    return convertBytesToGB(this.logstashPipelines.general.jvm.mem.nonHeapUsedInBytes);
  }

  get memoryTotal(): number {
    if (
      !this.logstashPipelines ||
      !this.logstashPipelines.general ||
      !this.logstashPipelines.general.jvm ||
      !this.logstashPipelines.general.jvm.mem
    ) {
      return 0;
    }

    return convertBytesToGB(this.logstashPipelines.general.jvm.mem.nonHeapCommittedInBytes);
  }

  get heapUsed(): number {
    if (
      !this.logstashPipelines ||
      !this.logstashPipelines.general ||
      !this.logstashPipelines.general.jvm ||
      !this.logstashPipelines.general.jvm.mem
    ) {
      return 0;
    }

    return convertBytesToGB(this.logstashPipelines.general.jvm.mem.heapUsedInBytes);
  }

  get percentHeapUsed(): number {
    if (
      !this.logstashPipelines ||
      !this.logstashPipelines.general ||
      !this.logstashPipelines.general.jvm ||
      !this.logstashPipelines.general.jvm.mem
    ) {
      return 0;
    }

    return this.logstashPipelines.general.jvm.mem.heapUsedPercent;
  }

  get heapTotal(): number {
    if (
      !this.logstashPipelines ||
      !this.logstashPipelines.general ||
      !this.logstashPipelines.general.jvm ||
      !this.logstashPipelines.general.jvm.mem
    ) {
      return 0;
    }

    return convertBytesToGB(this.logstashPipelines.general.jvm.mem.heapMaxInBytes);
  }

  get workers(): number {
    if (
      !this.logstashPipelines ||
      !this.logstashPipelines.general ||
      !this.logstashPipelines.general.pipeline
    ) {
      return 0;
    }
    return this.logstashPipelines.general.pipeline.workers;
  }

  get batchSize(): number {
    if (
      !this.logstashPipelines ||
      !this.logstashPipelines.general ||
      !this.logstashPipelines.general.pipeline
    ) {
      return 0;
    }
    return this.logstashPipelines.general.pipeline.batchSize;
  }

  get batchDelay(): number {
    if (
      !this.logstashPipelines ||
      !this.logstashPipelines.general ||
      !this.logstashPipelines.general.pipeline
    ) {
      return 0;
    }
    return this.logstashPipelines.general.pipeline.batchDelay;
  }



}
