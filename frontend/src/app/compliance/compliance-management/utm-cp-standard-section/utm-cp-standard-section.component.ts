import {Component, Input, OnInit} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {CpStandardSectionBehavior} from '../../shared/behavior/cp-standard-section.behavior';
import {CpStandardBehavior} from '../../shared/behavior/cp-standard.behavior';
// tslint:disable-next-line:max-line-length
import {UtmCpStandardSectionCreateComponent} from '../../shared/components/utm-cp-standard-section-create/utm-cp-standard-section-create.component';
import {CpStandardSectionService} from '../../shared/services/cp-standard-section.service';
import {ComplianceStandardSectionType} from '../../shared/type/compliance-standard-section.type';
import {ComplianceStandardType} from '../../shared/type/compliance-standard.type';
import {UtmCpStandardSectionDeleteComponent} from './utm-cp-standard-section-delete/utm-cp-standard-section-delete.component';

@Component({
  selector: 'app-utm-cp-standard-section',
  templateUrl: './utm-cp-standard-section.component.html',
  styleUrls: ['./utm-cp-standard-section.component.scss']
})
export class UtmCpStandardSectionComponent implements OnInit {
  standard: ComplianceStandardType;
  @Input() manage: boolean;
  standardSections: ComplianceStandardSectionType[] = [];
  loadingTemplates = true;
  noMoreResult: boolean;
  page = 1;
  solution: string;
  loadingMore: false;

  constructor(private cpStandardSectionService: CpStandardSectionService,
              private cpStandardBehavior: CpStandardBehavior,
              public cpStandardSectionBehavior: CpStandardSectionBehavior,
              private modalService: NgbModal) {
  }

  ngOnInit() {
    this.cpStandardBehavior.$standard.subscribe(standard => {
      if (standard) {
        this.standardSections = [];
        this.standard = standard;
        this.getSections();
      } else {
        this.standardSections = [];
        this.loadingTemplates = false;
        this.cpStandardSectionBehavior.$standardSection.next(null);
      }
    });
    this.cpStandardSectionBehavior.$updateStandardSection.subscribe(value => {
      if (value) {
        this.getSections();
      }
    });
  }

  getSections() {
    const query = {
      page: this.page - 1,
      size: 1000,
      sort: 'id,asc',
      'standardId.equals': this.standard.id,
      'standardSectionName.contains': this.solution
    };
    this.cpStandardSectionService.query(query).subscribe(response => {
      this.standardSections = response.body;
      this.loadingTemplates = false;
      this.cpStandardSectionBehavior.$standardSection.next(this.standardSections[0]);
    });
  }

  onSearchFor($event: string) {
    this.solution = $event;
    this.getSections();
  }

  onScroll() {

  }

  generateReport(section: ComplianceStandardSectionType) {

  }

  editSection(section: ComplianceStandardSectionType) {
    const modal = this.modalService.open(UtmCpStandardSectionCreateComponent, {centered: true});
    modal.componentInstance.standardSection = section;
    modal.componentInstance.standardSectionSaved.subscribe(() => {
      this.getSections();
    });
  }

  deleteSection(section: ComplianceStandardSectionType) {
    const modal = this.modalService.open(UtmCpStandardSectionDeleteComponent, {centered: true});
    modal.componentInstance.standardSection = section;
    modal.componentInstance.standardSectionDelete.subscribe(() => {
      this.getSections();
    });
  }

  createStandardSection() {
    const modal = this.modalService.open(UtmCpStandardSectionCreateComponent, {centered: true});
    modal.componentInstance.standardId = this.standard.id;
    modal.componentInstance.standardSectionSaved.subscribe(() => {
      this.getSections();
    });
  }
}

