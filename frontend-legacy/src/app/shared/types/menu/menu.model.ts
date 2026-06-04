export interface IMenu {
  id?: number;
  name?: string;
  url?: string;
  authorities?: string[];
  parentId?: number;
  type?: number;
  parent?: IMenu;
  menuActive?: boolean;
  dashboardId?: number;
  position?: number;
  childrens?: IMenu[];
  actions?: IMenu[];
  menuAction?: boolean;
  menuIcon?: string;
}

export class Menu implements IMenu {
  constructor(public id?: number,
              public name?: string,
              public url?: string,
              public authorities?: string[],
              public parentId?: number,
              public type?: number,
              public parent?: IMenu,
              public menuActive?: boolean,
              public dashboardId?: number,
              public position?: number,
              public childrens?: IMenu[],
              public actions?: IMenu[],
              public menuAction?: boolean,
              public menuIcon?: string) {
    this.id = id ? id : null;
    this.name = name ? name : null;
    this.url = url ? url : null;
    this.authorities = authorities ? authorities : [];
    this.parentId = parentId ? parentId : null;
    this.type = type ? type : null;
    this.parent = parent ? parent : null;
    this.menuActive = menuActive ? menuActive : null;
    this.dashboardId = dashboardId ? dashboardId : null;
    this.position = position ? position : null;
    this.childrens = childrens ? childrens : null;
    this.actions = actions ? actions : null;
    this.menuAction = menuAction ? menuAction : null;
    this.menuIcon = menuIcon ? menuIcon : null;
  }
}
