import {Component, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {ADMIN_ROLE} from '../../../shared/constants/global.constant';
import {ActionInitParamsEnum, ActionInitParamsValueEnum} from '../../../shared/enums/action-init-params.enum';
import {CpStandardBehavior} from '../../shared/behavior/cp-standard.behavior';
import {UtmComplianceCreateComponent} from '../../shared/components/utm-compliance-create/utm-compliance-create.component';
import {ComplianceStandardType} from '../../shared/type/compliance-standard.type';
import {UtmCpExportComponent} from '../utm-cp-export/utm-cp-export.component';
import {UtmCpImportComponent} from '../utm-cp-import/utm-cp-import.component';

@Component({
  selector: 'app-cp-standard-management',
  templateUrl: './cp-standard-management.component.html',
  styleUrls: ['./cp-standard-management.component.scss']
})
export class CpStandardManagementComponent implements OnInit {
  standard: ComplianceStandardType;
  admin = ADMIN_ROLE;

  constructor(private modalService: NgbModal,
              private activatedRoute: ActivatedRoute,
              private cpStandardBehavior: CpStandardBehavior) {
  }

  ngOnInit() {
    this.activatedRoute.queryParams.subscribe(params => {
      if (params[ActionInitParamsEnum.ON_INIT_ACTION]
        && params[ActionInitParamsEnum.ON_INIT_ACTION] === ActionInitParamsValueEnum.SHOW_CREATE_MODAL) {
        this.newCompliance();
      }
    });
  }

  onStandardChange($event: ComplianceStandardType) {
    this.standard = $event;
  }

  exportCompliance() {
    this.modalService.open(UtmCpExportComponent, {centered: true});
  }

  importCompliance() {
    const modalImport = this.modalService.open(UtmCpImportComponent, {centered: true});
    modalImport.componentInstance.reportsImported.subscribe(() => {
      this.cpStandardBehavior.$updateStandard.next(true);
    });
  }

  newCompliance() {
    this.modalService.open(UtmComplianceCreateComponent, {centered: true});
  }
}
