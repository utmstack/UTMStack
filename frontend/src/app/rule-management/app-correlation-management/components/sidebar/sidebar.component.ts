import {Component, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {ADMIN_ROLE} from '../../../../shared/constants/global.constant';

@Component({
  selector: 'app-rule-sidebar',
  templateUrl: './sidebar.component.html',
  styleUrls: ['./sidebar.component.scss']
})
export class SidebarComponent implements OnInit {
  adminAuth = ADMIN_ROLE;


  constructor(public router: Router,
              private spinner: NgxSpinnerService) {
  }

  ngOnInit() {}
  isActive(url): boolean {
    return this.router.isActive(url, false);
  }

  navigateTo(link: string) {
    this.spinner.show('loadingSpinner');
    this.router.navigate([link]).then(() => {
      this.spinner.hide('loadingSpinner');
    });
  }
}
