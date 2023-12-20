import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {Router} from '@angular/router';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';

@Component({
  selector: 'app-dashboard-export-preview',
  templateUrl: './dashboard-export-preview.component.html',
  styleUrls: ['./dashboard-export-preview.component.scss']
})
export class DashboardExportPreviewComponent implements OnInit {
  @Input() pdf: any;
  @Input() dashboardName: string;
  @Input() dashboardId: string;
  @Output() savePdf = new EventEmitter<boolean>();
  @Output() closePreview = new EventEmitter<boolean>();

  constructor(public activeModal: NgbActiveModal, private router: Router) {
  }

  ngOnInit() {
  }

  navigateToCustom() {
    this.closePreview.emit(true);
    let dashboard = this.dashboardName;
    dashboard = dashboard.replace(/\W+(?!$)/g, '-').toLowerCase();
    dashboard = dashboard.replace(/\W$/, '').toLowerCase();
    this.router.navigate(['/dashboard/customize-export/' + this.dashboardId + '/' + dashboard]).then(value =>
      this.activeModal.close());
  }

  save() {
    this.savePdf.emit(true);
    this.activeModal.close();
  }

  onCloseModal($event: boolean) {
    this.closePreview.emit(true);
  }
}
