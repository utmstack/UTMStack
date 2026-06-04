import {Component, Input, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {of} from "rxjs";
import {catchError, tap} from 'rxjs/operators';
import {ALERT_TIMESTAMP_FIELD} from '../../../../../shared/constants/alert/alert-field.constant';
import {LOG_ROUTE} from '../../../../../shared/constants/app-routes.constant';
import {LOG_INDEX_PATTERN, LOG_INDEX_PATTERN_ID} from '../../../../../shared/constants/main-index-pattern.constant';
import {ElasticOperatorsEnum} from '../../../../../shared/enums/elastic-operators.enum';
import {ElasticDataService} from '../../../../../shared/services/elasticsearch/elastic-data.service';

const LOG_ID_FIELD = 'id';

@Component({
  selector: 'app-alert-logs-related-action',
  templateUrl: './alert-logs-related-button.component.html',
  styleUrls: ['./alert-logs-related-button.component.css']
})
export class AlertLogsRelatedButtonComponent implements OnInit {

  @Input() logs: any[] = [];
  @Input() template: 'btn' | 'span'  = 'btn';
  showButton = false;

  constructor(private router: Router,
              private spinner: NgxSpinnerService,
              private elasticDataService: ElasticDataService) { }

  ngOnInit() {
    const ids = this.logs.map(log => log.id);
    const filters = [
      {field: LOG_ID_FIELD, operator: ElasticOperatorsEnum.CONTAIN_ONE_OF, value: ids},
      {field: ALERT_TIMESTAMP_FIELD, operator: ElasticOperatorsEnum.IS_BETWEEN, value: ['now-1y', 'now']}
    ];


    this.elasticDataService.exists(LOG_INDEX_PATTERN, filters).pipe(
      tap((exists: boolean) => this.showButton = exists),
      catchError(err => {
        console.error('Error checking related logs:', err);
        this.showButton = false;
        return of(false);
      })
    ).subscribe();

  }

  navigateToEvents() {
    const queryParams = {patternId: LOG_INDEX_PATTERN_ID, indexPattern: LOG_INDEX_PATTERN};
    queryParams[LOG_ID_FIELD] = ElasticOperatorsEnum.IS_ONE_OF + '->' + this.logs.map(log => log.id).slice(0, 100);
    queryParams[ALERT_TIMESTAMP_FIELD] = ElasticOperatorsEnum.IS_BETWEEN + '->' + 'now-1y' + ',' + 'now';
    this.spinner.show('loadingSpinner');
    this.router.navigate([LOG_ROUTE], {
      queryParams
    }).then(() => {
      this.spinner.hide('loadingSpinner');
    });
  }
}
