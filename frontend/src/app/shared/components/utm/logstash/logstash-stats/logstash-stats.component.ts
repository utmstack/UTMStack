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
    const part = this.logstashPipelines.general.jvm.mem.nonHeapUsedInBytes;
    const total = this.logstashPipelines.general.jvm.mem.nonHeapCommittedInBytes;
    const percentage = (part / total) * 100;
    return parseFloat(percentage.toFixed(2));
  }

  get memoryUsed(): number {
    return convertBytesToGB(this.logstashPipelines.general.jvm.mem.nonHeapUsedInBytes);
  }

  get memoryTotal(): number {
    return convertBytesToGB(this.logstashPipelines.general.jvm.mem.nonHeapCommittedInBytes);
  }

  get heapUsed(): number {
    return convertBytesToGB(this.logstashPipelines.general.jvm.mem.heapUsedInBytes);
  }

  get heapTotal(): number {
    return convertBytesToGB(this.logstashPipelines.general.jvm.mem.heapMaxInBytes);
  }

}
