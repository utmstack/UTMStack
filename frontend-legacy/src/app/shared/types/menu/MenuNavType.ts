import {Menu} from './menu.model';

export class MenuNavType {
  parent: Menu;
  childs: Menu[] = [];

  constructor(parent?: Menu, childs?: Menu[]) {
    this.parent = parent;
    this.childs = childs;
  }
}
