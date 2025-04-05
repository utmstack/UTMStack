import {HttpResponse} from '@angular/common/http';
import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {CpStandardBehavior} from '../../shared/behavior/cp-standard.behavior';
import {UtmCpStandardCreateComponent} from '../../shared/components/utm-cp-standard-create/utm-cp-standard-create.component';
import {CpStandardService} from '../../shared/services/cp-standard.service';
import {ComplianceStandardType} from '../../shared/type/compliance-standard.type';
import {UtmCpExportComponent} from '../utm-cp-export/utm-cp-export.component';
import {UtmCpStandardDeleteComponent} from './utm-cp-standard-delete/utm-cp-standard-delete.component';

@Component({
  selector: 'app-utm-cp-standard',
  templateUrl: './utm-cp-standard.component.html',
  styleUrls: ['./utm-cp-standard.component.scss']
})
export class UtmCpStandardComponent implements OnInit {
  standards: ComplianceStandardType[] = [];
  loading = true;
  @Input() standardId: number;
  @Output() standardChange = new EventEmitter<ComplianceStandardType>();
  @Input() manage: boolean;

  constructor(private cpStandardService: CpStandardService,
              private modalService: NgbModal,
              private cpStandardBehavior: CpStandardBehavior) {
  }

  ngOnInit() {
    this.getStandardList();
    this.cpStandardBehavior.$updateStandard.subscribe(value => {
      if (value) {
        this.getStandardList();
      }
    });
  }

  getStandardList() {
    this.cpStandardService.query({page: 0, size: 1000}).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  changeStandard(st: ComplianceStandardType) {
    this.standardId = st.id;
    this.standardChange.emit(st);
    this.cpStandardBehavior.$standard.next(st);
  }

  createStandard() {
    const modal = this.modalService.open(UtmCpStandardCreateComponent, {centered: true});
    modal.componentInstance.standardSaved.subscribe(() => {
      this.getStandardList();
    });
  }

  editStandard(standard: ComplianceStandardType) {
    const modal = this.modalService.open(UtmCpStandardCreateComponent, {centered: true});
    modal.componentInstance.standard = standard;
    modal.componentInstance.standardSaved.subscribe(() => {
      this.getStandardList();
    });
  }

  deleteStandard(standard: ComplianceStandardType) {
    const modal = this.modalService.open(UtmCpStandardDeleteComponent, {centered: true});
    modal.componentInstance.standard = standard;
    modal.componentInstance.standardDelete.subscribe(() => {
      this.getStandardList();
    });
  }

  private onSuccess(data, headers) {
    this.standards = data;
    this.loading = false;
    if (!this.loading) {
      if (this.standards.length > 0) {
        if (!this.standardId) {
          this.standardChange.emit(this.standards[0]);
          this.standardId = this.standards[0].id;
          this.cpStandardBehavior.$standard.next(this.standards[0]);
        } else {
          const standard = this.standards[this.standards.findIndex(value => value.id === this.standardId)];
          this.standardChange.emit(standard);
          this.cpStandardBehavior.$standard.next(standard);
        }
      } else {
        this.cpStandardBehavior.$standard.next(null);
        this.standardChange.emit(null);
      }
    }
  }

  private onError(body: any) {
  }

  exportStandard(st: ComplianceStandardType) {
    const modal = this.modalService.open(UtmCpExportComponent, {centered: true});
    modal.componentInstance.standardId = st.id;
  }
}
