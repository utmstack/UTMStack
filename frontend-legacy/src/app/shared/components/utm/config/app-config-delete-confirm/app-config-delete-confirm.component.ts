import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {SectionConfigType} from '../../../../types/configuration/section-config.type';

@Component({
  selector: 'app-app-config-delete-confirm',
  templateUrl: './app-config-delete-confirm.component.html',
  styleUrls: ['./app-config-delete-confirm.component.css']
})
export class AppConfigDeleteConfirmComponent implements OnInit {
  @Output() acceptDelete = new EventEmitter<boolean>();
  @Input() section: SectionConfigType;

  constructor(public activeModal: NgbActiveModal) {
  }

  ngOnInit() {
  }

  deleteSection() {
    this.acceptDelete.emit(true);
    this.activeModal.close();
  }
}
