import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {ICONMOON} from '../../../../constants/iconmoon.const';

@Component({
  selector: 'app-icon-select',
  templateUrl: './icon-select.component.html',
  styleUrls: ['./icon-select.component.scss']
})
export class IconSelectComponent implements OnInit {
  @Output() iconChange = new EventEmitter<string>();
  icons = ICONMOON;
  iconSelected: string;
  iconSearch: string;

  constructor(public activeModal: NgbActiveModal) {
  }

  ngOnInit() {
  }

  selectIcon(icon: string) {
    this.iconSelected = icon;
  }

  emitIconSelect() {
    this.iconChange.emit(this.iconSelected);
    this.activeModal.close();
  }

  searchIcon($event: string) {
    if (!$event || $event === '') {
      this.icons = ICONMOON;
    } else {
      this.icons = ICONMOON.filter(value => value.includes($event.toLowerCase()));
    }
  }
}
