import {Component, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {NgxSpinnerService} from 'ngx-spinner';
import {AccountService} from '../../../core/auth/account.service';
import {User} from '../../../core/user/user.model';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {ModalConfirmationComponent} from '../../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {LOG_ANALYZER_TOTAL_ITEMS} from '../../../shared/constants/log-analyzer.constant';
import {ITEMS_PER_PAGE} from '../../../shared/constants/pagination.constants';
import {SortEvent} from '../../../shared/directives/sortable/type/sort-event';
import {DataNatureTypeEnum} from '../../../shared/enums/nature-data.enum';
import {ElasticDataExportService} from '../../../shared/services/elasticsearch/elastic-data-export.service';
import {AlertReportType} from '../../../shared/types/alert/alert-report.type';
import {ElasticFilterType} from '../../../shared/types/filter/elastic-filter.type';
import {SortByType} from '../../../shared/types/sort-by.type';
import {AlertTagService} from '../shared/services/alert-tag.service';
import {AlertReportFilterComponent} from './shared/components/alert-report-filter/alert-report-filter.component';
import {AlertReportService} from './shared/services/report.service';

@Component({
  selector: 'app-alert-saved-reports',
  templateUrl: './alert-reports.component.html',
  styleUrls: ['./alert-reports.component.scss']
})
export class AlertReportsComponent implements OnInit {
  reports: AlertReportType[] = [];
  fields: SortByType[] = [
    {
      field: 'id',
      fieldName: 'Default'
    },
    {
      field: 'repName',
      fieldName: 'Name'
    },
    {
      field: 'repDate',
      fieldName: 'Date'
    }
  ];
  totalItems: any;
  page = 1;
  itemsPerPage = ITEMS_PER_PAGE;
  sort: string;
  filters: ElasticFilterType[];
  report: AlertReportType;
  user: User;
  loading = true;
  sortEvent: any;
  searching = false;
  search: string;
  generateReport: boolean;

  constructor(private reportService: AlertReportService,
              private toastService: UtmToastService,
              private modalService: NgbModal,
              private elasticDataExportService: ElasticDataExportService,
              private tagService: AlertTagService,
              private accountService: AccountService,
              public router: Router,
              private spinner: NgxSpinnerService) {
  }

  ngOnInit() {
    this.getAccount();
  }

  getAccount() {
    this.accountService.identity().then(account => {
      this.user = account;
      this.getReports();
    });
  }

  getReports() {
    const query = {
      sort: this.sort,
      page: this.page - 1,
      size: this.itemsPerPage,
      'userId.equals': this.user.id,
      'repName.contains': this.search
    };
    this.reportService.query(query).subscribe(res => {
      this.loading = false;
      this.reports = res.body;
      this.searching = false;
      this.totalItems = Number(res.headers.get('X-Total-Count'));
    });
  }

  openDeleteConfirmation(report) {
    this.report = report;
    const deleteModalRef = this.modalService.open(ModalConfirmationComponent, {backdrop: 'static', centered: true});
    deleteModalRef.componentInstance.header = 'Confirm delete operation';
    deleteModalRef.componentInstance.message = 'Are you sure you want to delete the report: ' + report.repName;
    deleteModalRef.componentInstance.confirmBtnText = 'Delete';
    deleteModalRef.componentInstance.confirmBtnIcon = 'icon-database-remove';
    deleteModalRef.componentInstance.confirmBtnType = 'delete';
    deleteModalRef.result.then(() => {
      this.deleteReport();
    });
  }

  deleteReport() {
    this.reportService.delete(this.report.id).subscribe(() => {
      const index = this.reports.findIndex(value => value.id === this.report.id);
      this.reports.splice(index, 1);
      this.toastService.showSuccessBottom('Report ' + this.report.repName + ' deleted successfully');
    });
  }

  viewFilter(report: AlertReportType) {
    this.report = report;
    const reportFilterModal = this.modalService.open(AlertReportFilterComponent, {centered: true});
    reportFilterModal.componentInstance.report = report;
  }

  onSortBy($event: SortEvent) {
    this.sortEvent = $event;
    this.sort = $event.column + ',' + $event.direction;
    this.getReports();
  }

  downloadToCsv(report: AlertReportType) {
    this.generateReport = true;
    const params = {
      columns: report.columns,
      dataOrigin: DataNatureTypeEnum.ALERT,
      filters: report.filters,
      top: LOG_ANALYZER_TOTAL_ITEMS
    };
    this.elasticDataExportService.exportCsv(params, 'UTM ALERTS').then(() => {
      this.generateReport = false;
    });
  }

  loadPage(page: number) {
    this.page = page;
    this.getReports();
  }

  onItemsPerPageChange($event: number) {
    this.itemsPerPage = $event;
    this.getReports();
  }

  onSearch($event: string) {
    this.search = $event;
    this.searching = true;
    this.getReports();
  }

  viewReport(report: AlertReportType) {
    this.spinner.show('loadingSpinner');
    const reportName = report.repName.toLowerCase().replace(' ', '_');
    this.router.navigate(['/data/alert/report/view/' + report.id + '/' + reportName]).then(() => {
      this.spinner.hide('loadingSpinner');
    });
  }
}
