import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {NavBehavior} from '../../../shared/behaviors/nav.behavior';
import {MenuService} from '../../../shared/services/menu/menu.service';
import {UtmDashboardService} from '../shared/services/utm-dashboard.service';

@Component({
  selector: 'app-dashboard-delete',
  templateUrl: './dashboard-delete.component.html',
  styleUrls: ['./dashboard-delete.component.scss']
})
export class DashboardDeleteComponent implements OnInit {
  @Input() dashboard: any;
  @Output() dashboardDeleted = new EventEmitter<string>();

  constructor(public activeModal: NgbActiveModal,
              private dashboardService: UtmDashboardService,
              private utmToastService: UtmToastService,
              private navBehavior: NavBehavior,
              private menuService: MenuService) {
  }

  ngOnInit() {
  }

  deleteDashboard() {
    this.dashboardService.delete(this.dashboard.id)
      .subscribe(() => {
        this.utmToastService.showSuccessBottom('Dashboard deleted successfully');
        this.activeModal.close();
        this.navBehavior.$nav.next(true);
        this.dashboardDeleted.emit('deleted');
      }, () => {
        this.utmToastService.showError('Error deleting dashboard',
          'Error deleting dashboard, please check your network and try again');
      });
  }
}
