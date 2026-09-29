import {Component} from '@angular/core';
import {Router} from '@angular/router';
import {SYSTEM_MENU_ICONS_PATH} from '../../../shared/constants/menu_icons.constants';

@Component({
  selector: 'app-federation-sidebar',
  templateUrl: './federation-sidebar.component.html',
  styleUrls: ['./federation-sidebar.component.scss']
})
export class FederationSidebarComponent {
  iconPath = SYSTEM_MENU_ICONS_PATH;

  constructor(private router: Router) {}

  isActive(link: string): boolean {
    return this.router.url.startsWith(link);
  }
}
