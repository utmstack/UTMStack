import {HttpResponse} from '@angular/common/http';
import {Component, Input, OnDestroy, OnInit} from '@angular/core';
import {LOG_INDEX_PATTERN, SOC_AI_INDEX_PATTERN} from '../../../../../shared/constants/main-index-pattern.constant';
import {ElasticOperatorsEnum} from '../../../../../shared/enums/elastic-operators.enum';
import {ElasticDataService} from '../../../../../shared/services/elasticsearch/elastic-data.service';
import {ElasticFilterType} from '../../../../../shared/types/filter/elastic-filter.type';
import {IndexSocAiStatus, SocAiType} from './soc-ai.type';

@Component({
  selector: 'app-alert-soc-ai',
  templateUrl: './alert-soc-ai.component.html',
  styleUrls: ['./alert-soc-ai.component.css']
})
export class AlertSocAiComponent implements OnInit, OnDestroy {
  @Input() alertID: string;
  @Input() socAiActive: boolean;
  socAiResponse: SocAiType;
  indexSocAiStatus = IndexSocAiStatus;
  loading = false;
  private interval: any;

  constructor(private elasticDataService: ElasticDataService, ) {}

  ngOnInit() {
    if (this.socAiActive) {
      this.getSocAiResponse();
    }
  }

  getSocAiResponse() {
    const filter: ElasticFilterType[] = [{
      field: 'activityId',
      operator: ElasticOperatorsEnum.IS,
      value: this.alertID
    }];
    this.elasticDataService.search(1, 1, 1, SOC_AI_INDEX_PATTERN, filter)
      .subscribe((res: HttpResponse<any>) => {
        this.loading = false;
        res.body.status = 'Completed';
        if (!res || res.body.status === IndexSocAiStatus.Processing) {
          if (!this.interval) {
            this.startInterval();
          }
        } else {
          // this.socAiResponse = res.body[0];
          this.socAiResponse = {
            "@timestamp": "2024-02-19T09:35:50.275613061Z",
            "severity": "3",
            status : 'Error',
            "category": "Brute Force",
            "alertName": "Attempts to Brute Force a Microsoft 365 User Account",
            "activityId": "b79a8f33-e9cd-4220-a270-cab5e523572b",
            "classification": "possible incident",
            "reasoning": [
              "The alert indicates that there have been failed login attempts to a Microsoft 365 user account using the same username 'jhondoe@gmail.com', originating from different IP addresses, which suggests a potential brute force attack. The logs confirm multiple failed login attempts from various IP addresses within a short time frame, triggering the alert for 'Attempts to Brute Force a Microsoft 365 User Account.' The log entries show consecutive failed login attempts due to 'IdsLocked' from different ClientIP addresses, indicating an automated process attempting unauthorized access. This pattern aligns with known brute force attack techniques."
            ],
            "nextSteps": []
          };
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

  ngOnDestroy() {
    clearInterval(this.interval);
  }

  processAlert() {

  }
}
