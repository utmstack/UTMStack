import {Component, EventEmitter, Input, OnDestroy, OnInit, Output} from '@angular/core';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {Subject} from 'rxjs';
import {takeUntil} from 'rxjs/operators';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {DashboardBehavior} from '../../../../shared/behaviors/dashboard.behavior';
import {UtmDashboardType} from '../../../../shared/chart/types/dashboard/utm-dashboard.type';
import {
  DashboardFilterLayout
} from '../../../../shared/components/utm/filters/dashboard-filter-view/dashboard-filter-view.component';
import {
  ElasticFilterDefaultTime
} from '../../../../shared/components/utm/filters/elastic-filter-time/elastic-filter-time.component';
import {ChangeFilterValueService} from '../../../../shared/components/utm/filters/services/change-filter-value.service';
import {FILTER_OPERATORS} from '../../../../shared/constants/filter-operators.const';
import {ElasticOperatorsEnum} from '../../../../shared/enums/elastic-operators.enum';
import {DataNatureTypeEnum, NatureDataPrefixEnum} from '../../../../shared/enums/nature-data.enum';
import {DashboardFilterType} from '../../../../shared/types/filter/dashboard-filter.type';
import {ElasticFilterType} from '../../../../shared/types/filter/elastic-filter.type';
import {OperatorsType} from '../../../../shared/types/filter/operators.type';
import {UtmIndexPattern} from '../../../../shared/types/index-pattern/utm-index-pattern';
import {UtmFieldType} from '../../../../shared/types/table/utm-field.type';
import {filtersWithPatternToStringParam} from '../../../../shared/util/query-params-to-filter.util';
import {CpReportBehavior} from '../../behavior/cp-report.behavior';
import {ComplianceScheduleService} from '../../services/compliance-schedule.service';
import {ComplianceReportType} from '../../type/compliance-report.type';
import {ComplianceScheduleFilterType} from '../../type/compliance-schedule-filter.type';
import {ComplianceScheduleType} from '../../type/compliance-schedule.type';

@Component({
  selector: 'app-utm-compliance-schedule-create',
  templateUrl: './utm-compliance-schedule-create.component.html',
  styleUrls: ['./utm-compliance-schedule-create.component.scss']
})
export class UtmComplianceScheduleCreateComponent implements OnInit, OnDestroy {
  @Input() report: ComplianceScheduleType;
  @Output() reportCreated = new EventEmitter<string>();
  @Output() reportUpdated = new EventEmitter<string>();
  dataNature: DataNatureTypeEnum = DataNatureTypeEnum.EVENT;
  step = 1;
  stepCompleted: number[] = [];
  creating = false;
  viewSection = false;
  standardSectionId: number;
  operators: OperatorsType[] = FILTER_OPERATORS;
  operatorEnum = ElasticOperatorsEnum;
  solution = '';
  cron = '0 0 0 */1 * *';
  reportId: number;
  filters: DashboardFilterType[];
  filtersTypes: ElasticFilterType[] = [];
  page = 1;
  private sortBy = NatureDataPrefixEnum.TIMESTAMP + ',' + 'desc';
  patterns: UtmIndexPattern[];
  defaultTime: ElasticFilterDefaultTime = new ElasticFilterDefaultTime('now-24h', 'now');
  pattern: UtmIndexPattern;
  queryParams: any;
  fields: UtmFieldType[] = [];
  filterDef: ComplianceScheduleFilterType[];
  dashboard: UtmDashboardType;
  onDestroy$: Subject<void> = new Subject();
  dashboardFilterLayout = DashboardFilterLayout.ScheduleReportCompliance;

  constructor(private complianceScheduleService: ComplianceScheduleService,
              public activeModal: NgbActiveModal,
              private cpReportBehavior: CpReportBehavior,
              private utmToastService: UtmToastService,
              private changeFilterValueService: ChangeFilterValueService,
              private dashboardBehavior: DashboardBehavior,
              public modalService: NgbModal) {
  }

  ngOnInit() {
    const req = {
      page: 0,
      size: 1000,
      sort: 'id,asc',
      'isActive.equals': true,
    };

    if (this.report) {
      this.solution = this.report.compliance.configSolution;
      this.viewSection = true;
      this.standardSectionId = this.report.compliance.standardSectionId;
      this.reportId = this.report.id;
      this.cron = this.report.scheduleString;
      this.filterDef = this.report.filterDef;
      this.filterDef.forEach( f => this.addFilterType({indexPattern: f.indexPattern, filter: f.filterType}));
    }

    this.dashboardBehavior.$filterDashboard
      .pipe(takeUntil(this.onDestroy$))
      .subscribe(data => {
        if (data && this.step === 2) {
          this.addFilterType(data);
        }
      });
  }

  backStep() {
    this.step -= 1;
    this.stepCompleted.pop();
  }

  nextStep() {
    this.stepCompleted.push(this.step);
    this.step += 1;
    if (this.step === 2 && this.report) {
        this.getAllFilters().forEach(f => {
          this.changeFilterValueService.changeSelectedValue({field: f.field, value: f.value});
        });
    }
  }

  isCompleted(step: number) {
    return this.stepCompleted.findIndex(value => value === step) !== -1;
  }

  createCompliance() {
    filtersWithPatternToStringParam(this.filtersTypes).then(queryParams => {
      const params  = queryParams && queryParams !== '' ? `${queryParams}` : '';
      this.creating = true;
      const reportCompliance: ComplianceScheduleType = {
        complianceId: this.reportId,
        filterDef: this.convertToFilterDefs(),
        scheduleString: this.cron,
        urlWithParams: `/dashboard/export-compliance/${this.reportId}?${params}`
      };
      if (this.report) {
        reportCompliance.id = this.report.id;
        this.complianceScheduleService.update(reportCompliance)
          .pipe(takeUntil(this.onDestroy$))
          .subscribe(() => {
          this.utmToastService.showSuccessBottom('Compliance report  edited successfully');
          this.filtersTypes = [];
          this.activeModal.close();
          this.reportUpdated.emit('edited');
        }, error1 => {
          this.creating = false;
          this.utmToastService.showError('Error', 'Error editing compliance report');
        });
      } else {
        this.complianceScheduleService.create(reportCompliance)
          .pipe(takeUntil(this.onDestroy$))
          .subscribe(() => {
          this.utmToastService.showSuccessBottom('Compliance report  created successfully');
          this.filtersTypes = [];
          this.activeModal.close();
          this.cpReportBehavior.$reportUpdate.next('update');
          this.reportCreated.emit('created');
        }, error1 => {
          this.creating = false;
          this.utmToastService.showError('Error', 'Error creating compliance report');
        });
      }
    });
  }

  onDashboardSelected($event: ComplianceReportType) {
    this.dashboard = $event.associatedDashboard;
    this.filters = JSON.parse(this.dashboard.filters);
    this.reportId = $event.id;
  }

  private onError(res: any) {
    this.utmToastService.showErrorResponse('Error', res);
  }

  getAllFilters(): ElasticFilterType[] {
    return this.report.filterDef.reduce((allFilters: ElasticFilterType[], currentDef) => {
      return allFilters.concat(currentDef.filterType);
    }, []);
  }

  convertToFilterDefs() {
    const filterDefs: ComplianceScheduleFilterType[] = [];

    this.filtersTypes.forEach(filterType => {
      const existingFilterDef = filterDefs.find(def => def.indexPattern === filterType.pattern);

      if (existingFilterDef) {
        existingFilterDef.filterType.push(filterType);
      } else {
        filterDefs.push({
          indexPattern: filterType.pattern,
          filterType: [filterType]
        });
      }
    });

    return filterDefs;
  }

  addFilterType(filter: any) {
    if (this.filtersTypes.length > 0) {
      const filterType = this.filtersTypes.find(f =>
        f.field === filter.filter[0].field &&
        f.operator === filter.filter[0].operator &&
        f.pattern === filter.indexPattern &&
        f.value !== filter.filter[0].value);

      if (filterType) {
        filterType.value = filter.filter[0].value;
      } else {
        this.filtersTypes.push({
          pattern: filter.indexPattern,
          value: filter.filter[0].value,
          operator: filter.filter[0].operator,
          field: filter.filter[0].field
        });
      }
    } else {
      this.filtersTypes.push({
        pattern: filter.indexPattern,
        value: filter.filter[0].value,
        operator: filter.filter[0].operator,
        field: filter.filter[0].field
      });
    }
  }

  ngOnDestroy() {
    this.onDestroy$.next();
    this.onDestroy$.complete();
  }
}
