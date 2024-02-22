import {HttpResponse} from '@angular/common/http';
import {Component, Input, OnDestroy, OnInit} from '@angular/core';
import {UtmToastService} from '../../../../../shared/alert/utm-toast.service';
import {LOG_INDEX_PATTERN, SOC_AI_INDEX_PATTERN} from '../../../../../shared/constants/main-index-pattern.constant';
import {ElasticOperatorsEnum} from '../../../../../shared/enums/elastic-operators.enum';
import {ElasticDataService} from '../../../../../shared/services/elasticsearch/elastic-data.service';
import {ElasticFilterType} from '../../../../../shared/types/filter/elastic-filter.type';
import {AlertSocAiService} from '../../services/alert-soc-ai.service';
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
  loadingProcess = false;
  private interval: any;

  constructor(private elasticDataService: ElasticDataService,
              private alertSocAiService: AlertSocAiService,
              private utmToastService: UtmToastService) {}

  ngOnInit() {
    if (this.socAiActive) {
      this.getSocAiResponse();
    }
  }

  getSocAiResponse() {
    this.loading = true;
    const filter: ElasticFilterType[] = [{
      field: 'activityId',
      operator: ElasticOperatorsEnum.IS,
      value: this.alertID
    }];
    this.elasticDataService.search(1, 1, 1, SOC_AI_INDEX_PATTERN, filter)
      .subscribe((res: HttpResponse<any>) => {
        this.loading = false;
        if (!res || res.body.length === 0) {
          this.socAiResponse = res.body;
        } else {
          this.socAiResponse = res.body[0];
        }

        if (this.socAiResponse.status ===  IndexSocAiStatus.Processing) {
          if (!this.interval) {
            this.startInterval();
          }
        } else {
          if (this.interval) {
            clearInterval(this.interval);
          }
        }
      },
      (res: HttpResponse<any>) => {
        this.loading = false;
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
    this.loadingProcess = true;
    this.alertSocAiService.processAlertBySoc([this.alertID])
      .subscribe((res) => {
          setTimeout(() => {
            this.loadingProcess = false;
            this.getSocAiResponse();
          }, 3000);
      },
      (error) => {
        this.utmToastService.showError('Error', 'An error occurred while processing the alert. Please try again later.');
        this.loadingProcess = false;
      });
  }

  isEmpty(object: any) {
    return object && Object.keys(object).length === 0;
  }
}
