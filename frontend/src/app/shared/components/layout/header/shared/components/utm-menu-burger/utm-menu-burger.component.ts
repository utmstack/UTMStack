import {Component, OnInit} from '@angular/core';
import {MenuBehavior} from '../../../../../../behaviors/menu.behavior';
import {ADMIN_ROLE} from '../../../../../../constants/global.constant';

@Component({
  selector: 'app-utm-menu-burger',
  templateUrl: './utm-menu-burger.component.html',
  styleUrls: ['./utm-menu-burger.component.css']
})
export class UtmMenuBurgerComponent implements OnInit {
  roleAdmin = ADMIN_ROLE;
  menuActive = false;

  constructor(private menuBehavior: MenuBehavior) {
  }

  ngOnInit() {
    this.menuBehavior.$menu.subscribe(activated => {
      this.menuActive = activated;
    });
  }

  menuController() {
    this.menuActive = !this.menuActive;
    this.menuBehavior.$menu.next(this.menuActive);
  }

}
