import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmModuleType} from '../../type/utm-module.type';

@Component({
  selector: 'app-app-module-deactivate',
  templateUrl: './app-module-deactivate.component.html',
  styleUrls: ['./app-module-deactivate.component.scss']
})
export class AppModuleDeactivateComponent implements OnInit {
  @Input() module: UtmModuleType;
  @Output() disable = new EventEmitter<boolean>();

  constructor(public activeModal: NgbActiveModal,) {
  }

  ngOnInit() {
  }

  continue() {
    this.disable.emit(true);
    this.activeModal.close();
  }
}
