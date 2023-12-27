import {Component, OnDestroy, OnInit} from '@angular/core';

import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {Observable, of, Subject} from 'rxjs';
import {catchError, map, takeUntil, tap} from 'rxjs/operators';
import {EventDataTypeEnum} from '../../data-management/alert-management/shared/enums/event-data-type.enum';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {ADMIN_ROLE} from '../../shared/constants/global.constant';
import {ITEMS_PER_PAGE} from '../../shared/constants/pagination.constants';
import {SortEvent} from '../../shared/directives/sortable/type/sort-event';
import {ElasticFilterType} from '../../shared/types/filter/elastic-filter.type';
import {CpReportBehavior} from '../shared/behavior/cp-report.behavior';
import {CpStandardBehavior} from '../shared/behavior/cp-standard.behavior';
import {
    UtmComplianceScheduleCreateComponent
} from '../shared/components/utm-compliance-schedule-create/utm-compliance-schedule-create.component';
import {
    UtmComplianceScheduleDeleteComponent
} from '../shared/components/utm-compliance-schedule-delete/utm-compliance-schedule-delete.component';
import {ComplianceScheduleService} from '../shared/services/compliance-schedule.service';
import {ComplianceScheduleFilterType} from '../shared/type/compliance-schedule-filter.type';
import {ComplianceScheduleType} from '../shared/type/compliance-schedule.type';
import {ComplianceStandardType} from '../shared/type/compliance-standard.type';
import {CronDescriptionGeneratorService} from "../shared/services/cron-description-generator.service";


@Component({
    selector: 'app-compliance-schedule',
    templateUrl: './compliance-schedule.component.html',
    styleUrls: ['./compliance-schedule.component.scss']
})
export class ComplianceScheduleComponent implements OnInit, OnDestroy {
    standard: ComplianceStandardType;
    admin = ADMIN_ROLE;
    protected readonly ITEMS_PER_PAGE = ITEMS_PER_PAGE;
    private requestParams: any;
    private sortBy: SortEvent;
    searching: any;
    checkbox: any;
    schedules: any[] = [];
    loading = false;
    totalItems: any;
    page = 1;
    itemsPerPage = ITEMS_PER_PAGE;
    selected: number[] = [];
    protected readonly EventDataTypeEnum = EventDataTypeEnum;
    schedules$: Observable<ComplianceScheduleType[]>;
    destroy$: Subject<void> = new Subject<void>();

    constructor(private modalService: NgbModal,
                private cpStandardBehavior: CpStandardBehavior,
                private utmToastService: UtmToastService,
                private cpReportBehavior: CpReportBehavior,
                private complianceScheduleService: ComplianceScheduleService,
                private cronDescriptionGeneratorService: CronDescriptionGeneratorService) {
    }

    ngOnInit() {
        this.requestParams = {
            page: this.page - 1,
            size: this.itemsPerPage,
            sort: this.sortBy,
        };
        this.getComplianceScheduleList();

        this.cpReportBehavior.$reportUpdate
            .pipe(takeUntil(this.destroy$))
            .subscribe(update => {
                if (update) {
                    this.getComplianceScheduleList();
                }
            });
    }

    newCompliance() {
        this.modalService.open(UtmComplianceScheduleCreateComponent, {
            centered: true,
            size: 'lg',
            windowClass: 'cp-schedule-report'
        });
    }

    onSearch($event: string | number) {
        this.searching = true;

        this.requestParams.page = 0;
        this.requestParams['name.contains'] = $event;
        this.getComplianceScheduleList();
    }

    getComplianceScheduleList() {
        this.loading = true;
        this.schedules$ = this.complianceScheduleService.query(this.requestParams)
            .pipe(
                tap((res) => {
                    this.totalItems = res.headers.get('X-Total-Count');
                    this.schedules = res.body;
                    this.searching = false;
                    this.loading = false;
                }),
                map(response => response.body),
                catchError(() => {
                        this.utmToastService.showError('Error', '"An error occurred while loading report schedules');
                        return of([]);
                    }
                ));
    }

    loadPage(page: any) {
        this.requestParams.page = page - 1;
        this.getComplianceScheduleList();
    }

    private onError(res: any) {
        this.utmToastService.showErrorResponse('Error', res);
    }

    addToSelected(dashboard: any) {
    }

    isSelected(schedule: any) {
        return false;
    }

    viewSchedule(schedule: any) {
    }

    editSchedule(schedule: any) {
        const modal = this.modalService.open(UtmComplianceScheduleCreateComponent, {
            centered: true,
            size: 'lg',
            windowClass: 'cp-schedule-report'
        });
        modal.componentInstance.report = schedule;
        modal.componentInstance.reportUpdated.subscribe(updated => {
            this.getComplianceScheduleList();
        });
    }

    deleteSchedule(schedule: any) {
        const modal = this.modalService.open(UtmComplianceScheduleDeleteComponent, {centered: true});
        modal.componentInstance.complianceSchedule = schedule;
        modal.componentInstance.complianceScheduleDeleted.subscribe(deleted => {
            this.getComplianceScheduleList();
        });
    }

    getAllFilters(filters: ComplianceScheduleFilterType[]): ElasticFilterType[] {
        return filters.reduce((allFilters: ElasticFilterType[], currentDef) => {
            return allFilters.concat(currentDef.filterType);
        }, []);
    }

    getCronExpression(cron: string){
      return this.cronDescriptionGeneratorService.getDescription(cron);
    }

    onSort($event: SortEvent) {
        this.requestParams.sort = $event.column + ',' + $event.direction;
        this.getComplianceScheduleList();
    }

    ngOnDestroy() {
        this.destroy$.next();
        this.destroy$.complete();
    }
}
