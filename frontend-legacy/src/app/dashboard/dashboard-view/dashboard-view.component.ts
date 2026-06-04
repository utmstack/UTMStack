import {Component, OnInit} from '@angular/core';
import {DomSanitizer} from '@angular/platform-browser';
import {Router} from '@angular/router';
import {DashboardBehavior} from '../../shared/behaviors/dashboard.behavior';
import {Menu} from '../../shared/types/menu/menu.model';

@Component({
  selector: 'app-dashboard-view',
  templateUrl: './dashboard-view.component.html',
  styleUrls: ['./dashboard-view.component.scss']
})
export class DashboardViewComponent implements OnInit {
  height: number;
  menu: Menu;
  url: any;

  constructor(private dashboardBehavior: DashboardBehavior,
              private router: Router,
              private sanitizer: DomSanitizer) {
  }

  ngOnInit() {
    this.height = window.outerHeight;
    this.dashboardBehavior.$dashboard.asObservable().subscribe(menu => {
      if (menu) {
        this.menu = menu;
        if (this.menu.url) {
          this.url = this.sanitizer.bypassSecurityTrustResourceUrl(this.menu.url);
        }
      }
    });
  }
}
