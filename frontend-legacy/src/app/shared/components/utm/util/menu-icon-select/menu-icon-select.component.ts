import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {MENU_ICONS, MENU_ICONS_PATH} from '../../../../constants/menu_icons.constants';

@Component({
  selector: 'app-icon-select',
  templateUrl: './menu-icon-select.component.html',
  styleUrls: ['./menu-icon-select.component.scss']
})
export class MenuIconSelectComponent implements OnInit {
  @Output() iconChange = new EventEmitter<string>();
  icons = MENU_ICONS;
  iconSelected: string;
  iconPath = MENU_ICONS_PATH;

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
      this.icons = MENU_ICONS;
    } else {
      this.icons = MENU_ICONS.filter(value => value.includes($event.toLowerCase()));
    }
  }
}
