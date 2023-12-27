import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {CpStandardBehavior} from '../../behavior/cp-standard.behavior';
import {CpStandardService} from '../../services/cp-standard.service';
import {ComplianceStandardType} from '../../type/compliance-standard.type';
import {UtmCpStandardCreateComponent} from '../utm-cp-standard-create/utm-cp-standard-create.component';

@Component({
  selector: 'app-utm-cp-standard-select',
  templateUrl: './utm-cp-standard-select.component.html',
  styleUrls: ['./utm-cp-standard-select.component.scss']
})
export class UtmCpStandardSelectComponent implements OnInit {
  standards: ComplianceStandardType[];
  @Input() standardId: number;
  @Input() required: boolean;
  @Output() standardSelect = new EventEmitter<number>();

  constructor(private cpStandardService: CpStandardService,
              private modalService: NgbModal,
              private standardBehavior: CpStandardBehavior) {
  }

  ngOnInit() {
    this.getStandards();
  }

  getStandards() {
    this.cpStandardService.query(
      {
        page: 0,
        size: 1000,
        'standardId.equals': this.standardId
      })
      .subscribe((response) => {
        this.standards = response.body;
      });
  }

  onSelectChange($event) {
    if ($event) {
      this.standardId = $event.id;
    }
    this.standardSelect.emit($event);
    this.standardBehavior.$standard.next($event);
  }

  newStandard() {
    const modal = this.modalService.open(UtmCpStandardCreateComponent, {centered: true});
    modal.componentInstance.standardSaved.subscribe((standard) => {
      this.standardId = standard.id;
      this.standardSelect.emit(standard);
      this.standardBehavior.$standard.next(standard);
      this.getStandards();
    });
  }

}
