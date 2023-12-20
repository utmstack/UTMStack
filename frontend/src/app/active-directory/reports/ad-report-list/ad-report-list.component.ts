import {HttpResponse} from '@angular/common/http';
import {Component, OnInit} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {JhiParseLinks} from 'ng-jhipster';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {ITEMS_PER_PAGE} from '../../../shared/constants/pagination.constants';
import {SortEvent} from '../../../shared/directives/sortable/type/sort-event';
import {SortByType} from '../../../shared/types/sort-by.type';
import {TIME_RANGE_MILLISECONDS} from '../../shared/const/time-range-milliseconds.const';
import {AdReportTypeEnum} from '../../tracker/shared/enums/ad-report-type.enum';
import {AdReportDeleteComponent} from '../ad-report-delete/ad-report-delete.component';
import {AdReportService} from '../shared/services/ad-report.service';
import {AdReportType} from '../shared/type/ad-report.type';

@Component({
    selector: 'app-ad-reports-list',
    templateUrl: './ad-report-list.component.html',
    styleUrls: ['./ad-report-list.component.scss']
})
export class AdReportListComponent implements OnInit {
    fields: SortByType[] = [
        {
            fieldName: 'Name',
            field: 'name'
        },
        {
            fieldName: 'First run',
            field: 'type'
        },
        {
            fieldName: 'Next Run',
            field: 'insecureUse'
        },
        {
            fieldName: 'Period',
            field: 'login'
        },
        {
            fieldName: 'Duration',
            field: 'login'
        }
    ];
    adReports: AdReportType[] = [];
    loading = false;
    totalItems: any;
    page: any;
    itemsPerPage = ITEMS_PER_PAGE;
    error: any;
    success: any;
    routeData: any;
    links: any;
    predicate: any;
    previousPage: any;
    reverse: any;
    viewDetail = 0;
    searching = false;
    search: string;
    private sortBy: string;
    private requestParams: any;
    private generateReport = false;
    private adReportTypeEnum = AdReportTypeEnum;

    constructor(private modalService: NgbModal,
                private activatedRoute: ActivatedRoute,
                private router: Router,
                private parseLinks: JhiParseLinks,
                private toastService: UtmToastService,
                private adReportService: AdReportService) {
    }

    ngOnInit() {
        this.requestParams = {
            page: this.page - 1,
            size: this.itemsPerPage,
            sort: this.sortBy,
            'name.contains': this.search
        };
        this.getReports();
    }

    resolveTimeRange(milliseconds: number) {
        return TIME_RANGE_MILLISECONDS.filter(value => value.value === milliseconds)[0].key;
    }

    deleteReport(adReport: any) {
        const modal = this.modalService.open(AdReportDeleteComponent, {centered: true});
        modal.componentInstance.adReport = adReport;
        modal.componentInstance.adReportDeleted.subscribe(deleted => {
            this.getReports();
        });
    }

    onSortBy($event: SortEvent) {
        this.sortBy = $event.column + ',' + $event.direction;
        this.getReports();
    }


    getReports() {
        this.loading = true;
        this.adReportService.query(this.requestParams).subscribe(
            (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
            (res: HttpResponse<any>) => this.onError(res.body)
        );
    }

    loadPage(page: number) {
        this.requestParams.page = page - 1;
        this.getReports();
    }

    onReportsFilterChange($event: any) {
        Object.keys($event).forEach(key => {
            if ($event[key] !== '' && $event[key] !== null) {
                this.requestParams[key] = $event[key];
            } else {
                this.requestParams[key] = undefined;
            }
        });
        this.getReports();
    }

    runReport(report: AdReportType) {
        this.generateReport = true;
        if (report.type === this.adReportTypeEnum.ACTIVITY) {
            report = {
                limit: report.limit,
                objectsType: report.objectsType,
                type: report.type,
                from: new Date(new Date().getTime() - (report.schedule)),
                to: new Date()
            };
        } else {
            report = {
                objectsType: report.objectsType,
                type: report.type,
            };
        }
        this.adReportService.exportToPdf(report).subscribe((dat) => {
            this.generateReport = false;
            const data = new Blob([dat], {type: 'application/pdf'});
            if (data.size > 0) {
                const fileURL = window.URL.createObjectURL(data);
                window.open(fileURL);
            } else {
                this.toastService.showWarning('NO DATA FOUND', 'No data found for this report');
            }
        });
    }

    searchReport($event: string) {
        this.searching = true;
        this.search = $event;
        this.requestParams['name.contains'] = $event;
        this.getReports();
    }

    private onSuccess(data, headers) {
        this.links = this.parseLinks.parse(headers.get('link'));
        this.totalItems = headers.get('X-Total-Count');
        this.adReports = data;
        this.searching = false;
        this.loading = false;
    }

    private onError(error) {
        // this.alertService.error(error.error, error.message, null);
    }
}
