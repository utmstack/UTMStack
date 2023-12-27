import {Component, Input, OnInit} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../../../shared/alert/utm-toast.service';
import {LOG_ANALYZER_TOTAL_ITEMS} from '../../../../../../shared/constants/log-analyzer.constant';
import {ALERT_INDEX_PATTERN} from '../../../../../../shared/constants/main-index-pattern.constant';
import {ElasticDataExportService} from '../../../../../../shared/services/elasticsearch/elastic-data-export.service';
import {AlertReportType} from '../../../../../../shared/types/alert/alert-report.type';
import {ElasticFilterType} from '../../../../../../shared/types/filter/elastic-filter.type';
import {UtmFieldType} from '../../../../../../shared/types/table/utm-field.type';
import {EventDataTypeEnum} from '../../../../shared/enums/event-data-type.enum';
import {AlertReportService} from '../../services/report.service';


@Component({
  selector: 'app-save-report',
  templateUrl: './save-report.component.html',
  styleUrls: ['./save-report.component.scss']
})
export class SaveAlertReportComponent implements OnInit {
  @Input() filters: ElasticFilterType[];
  @Input() fields: UtmFieldType[];
  @Input() dataType: EventDataTypeEnum;
  saving = false;
  generateReport = false;
  limitRange = [50, 100, 200, 500, 1000, 2500];
  limit = 50;
  name: any;
  description: any;
  report: AlertReportType;

  // private user: User;

  constructor(public activeModal: NgbActiveModal,
              // private accountService: AccountService,
              private reportService: AlertReportService,
              private utmToastService: UtmToastService,
              private elasticDataExportService: ElasticDataExportService,
  ) {
  }

  ngOnInit() {
    // this.getAccount();
  }

  // getAccount() {
  //   this.accountService.identity().then(account => {
  //     this.user = account;
  //   });
  // }

  // saveReport() {
  //   this.saving = true;
  //   const report: AlertReportType = {
  //     repDate: new Date(),
  //     repDescription: this.description,
  //     filters: this.filters,
  //     columns: this.fields,
  //     repLimit: this.limit,
  //     repName: this.name,
  //     userId: this.user.id
  //   };
  //   this.reportService.create(report).subscribe((rep) => {
  //     this.saving = false;
  //     this.report = rep.body;
  //     this.utmToastService.showSuccessBottom('Report ' + this.name.toString().toUpperCase() + ' has benn saved to your reports');
  //   });
  // }

  applyLimit(limit) {
    this.limit = limit;
  }

  exportToCsv() {
    this.generateReport = true;
    const params = {
      columns: this.fields,
      indexPattern: ALERT_INDEX_PATTERN,
      filters: this.filters,
      top: LOG_ANALYZER_TOTAL_ITEMS
    };
    this.elasticDataExportService.exportCsv(params, 'UTM ALERTS').then(() => {
      this.generateReport = false;
    });
  }

}
