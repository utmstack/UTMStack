import {Component, Input, OnInit} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {MenuCreateComponent} from '../../../shared/components/utm/util/menu-create/menu-create.component';
import {MenuService} from '../../../shared/services/menu/menu.service';
import {Menu} from '../../../shared/types/menu/menu.model';
import {MenuDeleteDialogComponent} from '../menu-delete/menu-delete-dialog.component';

@Component({
  selector: 'app-menu-card',
  templateUrl: './menu-card.component.html',
  styleUrls: ['./menu-card.component.scss']
})
export class MenuCardComponent implements OnInit {

  @Input()
  menu: Menu;

  hide = false;


  constructor(private menuService: MenuService,
              private modalService: NgbModal) {
  }

  ngOnInit() {
  }

  deleteMenu(id: number) {
    const modalRef = this.modalService.open(MenuDeleteDialogComponent, {backdrop: 'static', centered: true});
    modalRef.componentInstance.menu = this.menu;
    modalRef.result.then(() => {
      this.menuService.delete(id).subscribe(() => {
        this.menuService.loadMenu.next(true);
        this.menuService.loadSubMenu.next(null);
      });
    });
  }

  addSubmenu() {
    const modalRef = this.modalService.open(MenuCreateComponent, {backdrop: 'static', centered: true});
    modalRef.componentInstance.level = 2;
    modalRef.componentInstance.btnText = 'Create submenu';
    modalRef.componentInstance.isCreate = true;
    modalRef.componentInstance.menu = new Menu();
    modalRef.componentInstance.menu.parentId = this.menu.id;
    modalRef.componentInstance.menu.type = 2;
  }

  updateMenu() {
    this.hide = !this.hide;
  }

}
