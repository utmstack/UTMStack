import {HttpResponse} from '@angular/common/http';
import {Component, OnInit} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {NgxSpinnerService} from 'ngx-spinner';
import {ADMIN_ROLE} from '../../shared/constants/global.constant';
import {ComplianceParamsEnum} from '../shared/enums/compliance-params.enum';
import {ComplianceTypeEnum} from '../shared/enums/compliance-type.enum';
import {CpReportsService} from '../shared/services/cp-reports.service';
import {CpStandardSectionService} from '../shared/services/cp-standard-section.service';
import {CpStandardService} from '../shared/services/cp-standard.service';
import {ComplianceReportType} from '../shared/type/compliance-report.type';
import {ComplianceStandardSectionType} from '../shared/type/compliance-standard-section.type';
import {ComplianceStandardType} from '../shared/type/compliance-standard.type';

@Component({
  selector: 'app-compliance-templates',
  templateUrl: './compliance-templates.component.html',
  styleUrls: ['./compliance-templates.component.scss']
})
export class ComplianceTemplatesComponent implements OnInit {
  reports: ComplianceReportType[] = [];
  loadingTemplates = true;
  noMoreResult: boolean;
  page = 1;
  loadingMore: boolean;
  solution: string;
  standardSectionId: number;
  admin = ADMIN_ROLE;
  standardId: number;
  standards: ComplianceStandardType[] = [];
  sections: ComplianceStandardSectionType[] = [];
  sectionId: number;
  loading: boolean;
  standard: ComplianceStandardType;
  section: ComplianceStandardSectionType;
  private standardName: string;
  showDetailFor = 0;

  detail: ComplianceReportType;


  constructor(private router: Router,
              private cpReportsService: CpReportsService,
              private activatedRoute: ActivatedRoute,
              private modalService: NgbModal,
              private spinner: NgxSpinnerService,
              private cpStandardService: CpStandardService,
              private cpStandardSectionService: CpStandardSectionService) {
  }

  ngOnInit() {
    this.activatedRoute.queryParams.subscribe(params => {
      if (params && params.standardId) {
        this.standardId = Number(params[ComplianceParamsEnum.STANDARD_ID]);
        this.sectionId = Number(params[ComplianceParamsEnum.SECTION_ID]);
        this.standardName = params[ComplianceParamsEnum.STANDARD_NAME];
      }
    });
    this.getStandardList();
  }

  generateReport(template?: ComplianceReportType) {
    const queryParams = {};
    queryParams[ComplianceParamsEnum.TEMPLATE] = template.id;
    queryParams[ComplianceParamsEnum.STANDARD_ID] = this.standardId;
    queryParams[ComplianceParamsEnum.SECTION_ID] = this.sectionId;
    const url = template.configType === ComplianceTypeEnum.TEMPLATE ? 'compliance/template-result' : 'compliance/template-custom';
    this.spinner.show('loadingSpinner');
    this.router.navigate([url], {
      queryParams
    }).then(() => this.spinner.hide('loadingSpinner'));
  }

  getSections(standardId: number) {
    const query = {
      page: this.page - 1,
      size: 1000,
      sort: 'id,asc',
      'standardId.equals': standardId
    };
    this.cpStandardSectionService.query(query).subscribe(response => {
      this.sections = response.body;
      if (!this.sectionId) {
        this.sectionId = this.sections[0].id;
      }
      this.section = this.sections[0];
      this.getTemplatesReports(standardId, this.sectionId);
      this.loading = false;
    });
  }

  getTemplatesReports(standardId: number, sectionId: number) {
    const query = {
      page: this.page - 1,
      size: 1000,
      sort: 'id,asc',
      standardId,
      solution: this.solution,
      sectionId
    };
    this.cpReportsService.queryByStandard(query).subscribe(response => {
      this.reports = [];
      this.reports = response.body;
      this.loadingTemplates = false;
    });
  }

  onSearchFor($event: string) {
    this.solution = $event;
    this.getTemplatesReports(this.standardId, this.sectionId);
  }

  onStandardSectionSelectChange(standardId: number, section: ComplianceStandardSectionType) {
    this.section = section;
    this.standardSectionId = section.id;
    this.sectionId = section.id;
    this.reports = [];
    this.getTemplatesReports(standardId, section.id);
  }

  changeStandard(st: ComplianceStandardType) {
    this.standardId = st.id;
    this.sectionId = undefined;
    this.standard = st;
    this.reports = [];
    this.getSections(this.standardId);
  }

  getStandardList() {
    this.cpStandardService.query({page: 0, size: 1000}).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  private onSuccess(data, headers) {
    this.standards = data;
    if (!this.standardId) {
      this.standardId = this.standards[0].id;
      this.standard = this.standards[0];
    } else {
      const stIndex = this.standards.findIndex(value => value.id === this.standardId);
      this.standard = this.standards[stIndex];
    }
    this.getSections(this.standardId);
  }

  private onError(body: any) {
  }

  viewSolution(report: ComplianceReportType): void {
    if (!this.detail) {
      this.detail = report;
    } else if (this.detail.id === report.id) {
      this.detail = undefined;
    } else {
      this.detail = report;
    }
  }
}
