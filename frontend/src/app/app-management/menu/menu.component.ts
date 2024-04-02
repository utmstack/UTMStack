import {Component, OnInit} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {DndDropEvent, DropEffect} from 'ngx-drag-drop';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {NavBehavior} from '../../shared/behaviors/nav.behavior';
import {MenuCreateComponent} from '../../shared/components/utm/util/menu-create/menu-create.component';
import {MenuService} from '../../shared/services/menu/menu.service';
import {CheckForUpdatesService} from '../../shared/services/updates/check-for-updates.service';
import {IMenu, Menu} from '../../shared/types/menu/menu.model';
import {QueryType} from '../../shared/types/query-type';
import {MenuDeleteDialogComponent} from './menu-delete/menu-delete-dialog.component';

interface MenuListItem {
  id?: number;
  name?: string;
  url?: string;
  authorities?: string[];
  parentId?: number;
  type?: number;
  parent?: IMenu;
  menuActive?: boolean;
  children?: MenuListItem[];
}

@Component({
  selector: 'app-menu',
  templateUrl: './menu.component.html',
  styleUrls: ['./menu.component.scss']
})
export class MenuComponent implements OnInit {
  private currentDraggableEvent: DragEvent;
  private currentDragEffectMsg: string;
  menus: Menu[];
  deployed: number[] = [];
  menuActive: any;
  loading = true;
  private isInDevelop: boolean;

  constructor(private menuService: MenuService,
              private modalService: NgbModal,
              private checkForUpdatesService: CheckForUpdatesService,
              private utmToastService: UtmToastService,
              private navBehavior: NavBehavior) {
  }

  ngOnInit() {
    this.getMode();
  }

  getMode() {
    this.checkForUpdatesService.getMode().subscribe(response => {
      this.isInDevelop = response.body;
      this.loadMenus();
    });
  }

  loadMenus() {
    this.loading = true;
    this.menus = [];
    const query = new QueryType();
    query.add('size', 5000);
    this.menuService.getMenuStructure(false).subscribe(reponse => {
      // Getting menu that are not from system
      if (this.isInDevelop) {
        this.menus = reponse.body;
      } else {
        this.menus = reponse.body.filter(value => value.type !== 1);
      }
      this.loading = false;
    });
  }

  deleteSubmenu(submenu: Menu) {
    const modalRef = this.modalService.open(MenuDeleteDialogComponent, {backdrop: 'static', centered: true});
    modalRef.result.then(() => {
      this.menuService.delete(submenu.id).subscribe(() => {
        this.loadMenus();
        this.navBehavior.$nav.next(true);
      });
    });
  }

  updateSubmenu(submenu: Menu) {
    const modalRef = this.modalService.open(MenuCreateComponent,
      {backdrop: 'static', centered: true});
    modalRef.componentInstance.level = 2;
    modalRef.componentInstance.btnText = 'Update submenu';
    modalRef.componentInstance.isCreate = false;
    modalRef.componentInstance.menu = submenu;
    if (submenu.parentId) {
      modalRef.componentInstance.parent = this.menus.find( m => m.id === submenu.parentId);
    }
    modalRef.componentInstance.menuCreated.subscribe(() => {
      this.loadMenus();
    });
  }

  addMenu() {
    const modalRef = this.modalService.open(MenuCreateComponent,
      {backdrop: 'static', centered: true});
    modalRef.componentInstance.level = 1;
    modalRef.componentInstance.isCreate = true;
    modalRef.componentInstance.isContainer = true;
    modalRef.componentInstance.menuCreated.subscribe(() => {
      this.navBehavior.$nav.next(true);
      this.loadMenus();
    });
  }

  addSubmenu(menu) {
    const modalRef = this.modalService.open(MenuCreateComponent, {backdrop: 'static', centered: true});
    modalRef.componentInstance.level = 2;
    modalRef.componentInstance.btnText = 'Create submenu';
    modalRef.componentInstance.isCreate = true;
    modalRef.componentInstance.menu = new Menu();
    modalRef.componentInstance.menu.parentId = menu.id;
    modalRef.componentInstance.parent = menu;
    modalRef.componentInstance.menu.menuActive = true;
    modalRef.componentInstance.menu.type = 2;
    modalRef.componentInstance.isContainer = false;
    modalRef.componentInstance.menuCreated.subscribe(() => {
      this.navBehavior.$nav.next(true);
      this.loadMenus();
    });
  }


  toggleActiveMenu(menu: Menu) {
    menu.menuActive = !menu.menuActive;
    if (menu.childrens) {
      menu.childrens.forEach(value => value.menuActive = menu.menuActive);
    }
  }

  isActiveMenu(menu: MenuListItem) {
  }

  isActiveMenuChildren(menuChildren: MenuListItem) {
  }

  addToView(item: any) {
    const indexDeployed = this.deployed.findIndex(value => value === item.id);
    if (indexDeployed === -1) {
      this.deployed.push(item.id);
    } else {
      this.deployed.splice(indexDeployed, 1);
    }
  }

  onDragged(item: Menu, list: any[], effect: DropEffect) {
    if (effect === 'move' && item.url) {
      const index = list.indexOf(item);
      list.splice(index, 1);
    }
  }

  onDragEnd(event: DragEvent) {
    this.currentDraggableEvent = event;
  }

  onDrop(event: DndDropEvent, list?: any[], menu?: Menu) {
    // console.log('drop ->>> ', event.data, menu);
    if (list && event.dropEffect === 'move' && event.data.url && !menu.url) {
      let index = event.index;
      if (typeof index === 'undefined') {
        index = list.length;
      }
      list.splice(index, 0, event.data);
    }
  }

  saveMenuStructure() {
    this.setMenuPosition().then(menus => {
      this.menuService.saveStructure(menus).subscribe(value => {
        this.navBehavior.$nav.next(true);
        this.utmToastService.showSuccessBottom('Menu edited successfully');
      });
    });
  }

  setMenuPosition(): Promise<Menu[]> {
    return new Promise<Menu[]>(resolve => {
      resolve(this.setPosition(this.menus));
    });
  }

  setPosition(menu: Menu[], parent?: Menu): Menu[] {
    menu.forEach((value, index) => {
      value.position = index;
      if (parent) {
        value.parentId = parent.id;
      }
      if (value.childrens) {
        this.setPosition(value.childrens, value);
      }
    });
    return menu;
  }
}
