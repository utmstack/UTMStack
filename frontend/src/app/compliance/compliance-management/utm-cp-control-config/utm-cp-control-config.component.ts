import {Component, isDevMode, OnInit} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {CpReportBehavior} from '../../shared/behavior/cp-report.behavior';
import {CpStandardSectionBehavior} from '../../shared/behavior/cp-standard-section.behavior';
import {
  UtmComplianceControlConfigCreateComponent
} from '../../shared/components/utm-compliance-control-config-create/utm-compliance-control-config-create.component';
import {CpControlConfigService} from '../../shared/services/cp-control-config.service';
import {ComplianceControlConfigType} from '../../shared/type/compliance-control-config.type';
import {ComplianceStandardSectionType} from '../../shared/type/compliance-standard-section.type';
import {UtmCpControlConfigDeleteComponent} from './utm-cp-control-config-delete/utm-cp-control-config-delete.component';

@Component({
  selector: 'app-utm-cp-control-config',
  templateUrl: './utm-cp-control-config.component.html',
  styleUrls: ['./utm-cp-control-config.component.scss']
})
export class UtmCpControlConfigComponent implements OnInit {
  section: ComplianceStandardSectionType;
  complianceControls: ComplianceControlConfigType[] = [];
  loadingTemplates = true;
  page = 1;
  solution: string;
  showDetailFor = 0;
  isDevMode = isDevMode;

  constructor(private cpControlConfigService: CpControlConfigService,
              public cpStandardSectionBehavior: CpStandardSectionBehavior,
              private cpReportBehavior: CpReportBehavior,
              private modalService: NgbModal) {
  }

  ngOnInit() {
    this.cpStandardSectionBehavior.$standardSection.subscribe(section => {
      if (section) {
        this.section = section;
        this.getControls();
      } else {
        this.complianceControls = [];
        this.section = null;
        this.loadingTemplates = false;
      }
    });

    this.cpReportBehavior.$reportUpdate.subscribe(update => {
      if (update) {
        this.getControls();
      }
    });

  }

  getControls() {
    const query = {
      page: this.page - 1,
      size: 1000,
      sort: 'id,asc',
      'standardSectionId.equals': this.section.id,
      'configSolution.contains': this.solution
    };
    this.complianceControls = [];
    this.cpControlConfigService.query(query).subscribe(response => {
      this.complianceControls = response.body;
      this.loadingTemplates = false;
    });
  }

  onSearchFor($event: string) {
    this.solution = $event;
    this.getControls();
  }

  deleteControl(control: ComplianceControlConfigType) {
    const modal = this.modalService.open(UtmCpControlConfigDeleteComponent, {centered: true});
    modal.componentInstance.control = control;
    modal.componentInstance.controlDelete.subscribe(() => {
      this.getControls();
    });
  }

  editControl(control: ComplianceControlConfigType) {
    const controlModal = this.modalService.open(UtmComplianceControlConfigCreateComponent, {centered: true, size: 'lg'});
    controlModal.componentInstance.control = control;
    controlModal.componentInstance.controlCreated.subscribe(() => {
      this.getControls();
    });
  }

  toggleDetail(id: number) {
    this.showDetailFor = this.showDetailFor === id ? 0 : id;
  }
}
