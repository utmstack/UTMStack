import {HttpResponse} from '@angular/common/http';
import {Component, Input, OnDestroy, OnInit} from '@angular/core';
import {LOG_INDEX_PATTERN, SOC_AI_INDEX_PATTERN} from '../../../../../shared/constants/main-index-pattern.constant';
import {ElasticOperatorsEnum} from '../../../../../shared/enums/elastic-operators.enum';
import {ElasticDataService} from '../../../../../shared/services/elasticsearch/elastic-data.service';
import {ElasticFilterType} from '../../../../../shared/types/filter/elastic-filter.type';
import {SocAiType} from './soc-ai.type';

@Component({
  selector: 'app-alert-soc-ai',
  templateUrl: './alert-soc-ai.component.html',
  styleUrls: ['./alert-soc-ai.component.css']
})
export class AlertSocAiComponent implements OnInit, OnDestroy {
  @Input() alertID: string;
  @Input() socAiActive: boolean;
  socAiResponse: SocAiType;
  loading = false;
  private interval: any;

  constructor(private elasticDataService: ElasticDataService, ) {
  }

  ngOnInit() {
    if (this.socAiActive) {
      this.getSocAiResponse();
    }
  }

  ngOnDestroy() {
    clearInterval(this.interval);
  }

  getSocAiResponse() {
    const filter: ElasticFilterType[] = [{
      field: 'activityId',
      operator: ElasticOperatorsEnum.IS,
      value: this.alertID
    }];
    this.elasticDataService.search(1, 1,
      1, SOC_AI_INDEX_PATTERN, filter).subscribe(
      (res: HttpResponse<any>) => {
        this.loading = false;
        if (!res || res.body === null || res.body.length === 0) {
          if (!this.interval) {
            this.startInterval();
          }
        } else {
          this.socAiResponse = res.body[0];
        }
      },
      (res: HttpResponse<any>) => {
      }
    );
  }

  startInterval() {
    this.interval = setInterval(() => {
      this.getSocAiResponse();
    }, 10000);
  }

}
