import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {CpStandardSectionBehavior} from '../../behavior/cp-standard-section.behavior';
import {CpStandardBehavior} from '../../behavior/cp-standard.behavior';
import {CpStandardSectionService} from '../../services/cp-standard-section.service';
import {ComplianceStandardSectionType} from '../../type/compliance-standard-section.type';
import {UtmCpStandardSectionCreateComponent} from '../utm-cp-standard-section-create/utm-cp-standard-section-create.component';

@Component({
  selector: 'app-utm-cp-st-section-select',
  templateUrl: './utm-cp-st-section-select.component.html',
  styleUrls: ['./utm-cp-st-section-select.component.scss']
})
export class UtmCpStSectionSelectComponent implements OnInit {
  standardsSection: ComplianceStandardSectionType[];
  @Input() standardSectionId: number;
  @Input() standardSectionEditId: number;
  @Input() required: boolean;
  @Input() onlyWithReport: boolean;
  @Output() standardSectionSelect = new EventEmitter<number>();
  private standarId: number;

  constructor(private cpStandardSectionService: CpStandardSectionService,
              private cpStandardBehavior: CpStandardBehavior,
              private cpStandardSectionBehavior: CpStandardSectionBehavior,
              private modalService: NgbModal) {
  }

  ngOnInit() {
    this.cpStandardBehavior.$standard.subscribe(value => {
      if (value) {
        this.standarId = value.id;
        this.standardSectionId = undefined;
        this.standardSectionSelect.emit(this.standardSectionId);
        this.getStandardSection(value);
      } else {
        this.standarId = null;
        this.standardSectionId = undefined;
        this.standardsSection = [];
      }
    });
    if (this.standardSectionEditId) {
      this.standardSectionId = this.standardSectionEditId;
      this.standardSectionSelect.emit(this.standardSectionId);
    }

  }

  getStandardSection(standard) {
    if (this.onlyWithReport) {
      this.getStandardsWithReports(standard.id);
    } else {
      this.getStandardsAllStandard(standard.id);
    }
  }

  getStandardsWithReports(standardId: number) {
    this.cpStandardSectionService.queryWithReports({page: 0, size: 1000, standardId})
      .subscribe((response) => {
        this.standardsSection = [];
        this.standardsSection = response.body;
      });
  }


  getStandardsAllStandard(standardId: number) {
    this.standardsSection = [];
    this.cpStandardSectionService.query({
      page: 0,
      size: 1000,
      'standardId.equals': standardId
    })
      .subscribe((response) => {
        this.standardsSection = response.body;
      });
  }

  onSelectChange($event) {
    this.standardSectionId = $event ? $event.id : null;
    this.standardSectionSelect.emit(this.standardSectionId);
  }

  newStandardSection() {
    const modal = this.modalService.open(UtmCpStandardSectionCreateComponent, {centered: true});
    modal.componentInstance.standardId = this.standarId;
    modal.componentInstance.standardSectionSaved.subscribe((section) => {
      this.standardSectionId = section.id;
      this.standardSectionSelect.emit(section.id);
      this.cpStandardSectionBehavior.$updateStandardSection.next(true);
      this.getStandardSection(this.standarId);
    });
  }
}
