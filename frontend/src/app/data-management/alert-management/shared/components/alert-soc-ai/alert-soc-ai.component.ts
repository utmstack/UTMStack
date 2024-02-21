import {HttpResponse} from '@angular/common/http';
import {Component, Input, OnDestroy, OnInit} from '@angular/core';
import {LOG_INDEX_PATTERN, SOC_AI_INDEX_PATTERN} from '../../../../../shared/constants/main-index-pattern.constant';
import {ElasticOperatorsEnum} from '../../../../../shared/enums/elastic-operators.enum';
import {ElasticDataService} from '../../../../../shared/services/elasticsearch/elastic-data.service';
import {ElasticFilterType} from '../../../../../shared/types/filter/elastic-filter.type';
import {IndexSocAiStatus, SocAiType} from './soc-ai.type';
import {AlertSocAiService} from "../../services/alert-soc-ai.service";

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

  constructor(private elasticDataService: ElasticDataService,
              private alertSocAiService: AlertSocAiService) {}

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
       // res.body.status = 'Completed';
        if (!res || res.body.length === 0) {
          this.socAiResponse = res.body;
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

  ngOnDestroy() {
    clearInterval(this.interval);
  }

  processAlert() {
    this.alertSocAiService.processAlertBySoc([this.alertID])
      .subscribe((res) => {
        console.log(res);
      },
      error => console.log(error));
  }

  isEmpty(object: any) {
    return Object.keys(object).length === 0;
  }
}
