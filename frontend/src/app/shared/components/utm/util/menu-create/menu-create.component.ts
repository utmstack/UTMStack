import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UserService} from '../../../../../core/user/user.service';
import {UtmToastService} from '../../../../alert/utm-toast.service';
import {NavBehavior} from '../../../../behaviors/nav.behavior';
import {DEMO_URL} from '../../../../constants/global.constant';
import {MENU_ICONS_PATH} from '../../../../constants/menu_icons.constants';
import {MenuService} from '../../../../services/menu/menu.service';
import {Menu} from '../../../../types/menu/menu.model';
import {QueryType} from '../../../../types/query-type';
import {ContactUsComponent} from '../../../contact-us/contact-us.component';
import {MenuIconSelectComponent} from '../menu-icon-select/menu-icon-select.component';

@Component({
  selector: 'app-menu-create',
  templateUrl: './menu-create.component.html',
  styleUrls: ['./menu-create.component.scss']
})
export class MenuCreateComponent implements OnInit {
  @Input() btnText = 'Add';
  @Input() isCreate = true;
  @Input() level = 1;
  @Input() menu: Menu;
  @Input() parent: Menu;
  @Input() isContainer = true;
  @Output() menuCreated = new EventEmitter<Menu>();
  authorities: string[];
  creating = false;
  isActive: boolean;
  icon: string;
  iconPath = MENU_ICONS_PATH;
  parentsMenu: Menu[];
  menuParentId: number;

  constructor(private menuService: MenuService,
              private modalService: NgbModal,
              private navBehavior: NavBehavior,
              private userService: UserService,
              public activeModal: NgbActiveModal,
              private utmToastService: UtmToastService) {
    this.userService.authorities().subscribe(result => {
      if (result) {
        this.authorities = result;
      }
    });
  }


  ngOnInit() {
    if (!this.menu) {
      this.menu = {};
      this.menu.menuAction = false;
      this.menu.menuActive = true;
      this.menu.authorities = [];
      if (this.level === 1) {
        this.menu.type = 2;
      }
    } else {

      this.isContainer = !this.parent;
      this.menu.menuActive = true;
      this.menuParentId = this.parent ? this.parent.id : null;
    }
    this.getParentMenu();
  }


  getParentMenu() {
    const query = new QueryType();
    query.add('parentId.specified', false);
    query.add('modulesNameShort.specified', false);
    this.menuService.query(query).subscribe(response => {
      this.parentsMenu = response.body;
    });
  }

  createMenu() {
    this.creating = true;
    if (!window.location.href.includes(DEMO_URL)) {
      if (!this.isCreate) {
        this.menuService.update(this.menu).subscribe((menu) => {
          this.creating = false;
          if (this.level === 1) {
            this.menu = {};
            this.menuService.loadMenu.next(true);
          } else if (this.level === 2) {
            this.modalService.dismissAll('ok');
            this.menuService.loadSubMenu.next(this.menu.parentId);
          }
          this.activeModal.close();
          this.menuCreated.emit(menu.body);
          this.navBehavior.$nav.next(true);
          this.utmToastService.showSuccessBottom('Menu edited successfully');
        });
      } else {
        this.menu.type = 2;
        this.menuService.create(this.menu).subscribe((menu) => {
          this.creating = false;
          this.menuCreated.emit(menu.body);
          this.navBehavior.$nav.next(true);
          if (this.level === 1) {
            this.menu = {};
            this.menuService.loadMenu.next(true);
          } else if (this.level === 2) {
            this.modalService.dismissAll('ok');
            this.menuService.loadSubMenu.next(this.menu.parentId);
          }
          this.activeModal.close();
          this.utmToastService.showSuccessBottom('Menu created successfully');
        },
          error => this.creating = false);
      }
    } else {
      this.modalService.open(ContactUsComponent, {centered: true});
    }
  }

  cancelOrClose() {
    this.modalService.dismissAll('cancel');
  }

  selectIcon() {
    const modal = this.modalService.open(MenuIconSelectComponent, {centered: true});
    modal.componentInstance.iconChange.subscribe(icon => {
      this.menu.menuIcon = icon;
    });
  }

  toggleMenuActive($event: boolean) {
    this.isContainer = $event;
    if (!this.isContainer) {
      this.menu.url = null;
    }
  }
}
