import {AfterViewChecked, ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {ResizeEvent} from 'angular-resizable-element';
import {AdReportCreateComponent} from '../../reports/ad-report-create/ad-report-create.component';
import {TreeObjectBehavior} from '../../shared/behavior/tree-object.behvior';
import {resolveType} from '../../shared/functions/ad-util.function';
import {ActiveDirectoryService} from '../../shared/services/active-directory.service';
import {ActiveDirectoryType} from '../../shared/types/active-directory.type';
import {AdTrackerCreateComponent} from '../../tracker/ad-tracker-create/ad-tracker-create.component';

@Component({
  selector: 'app-active-directory-view',
  templateUrl: './active-directory-view.component.html',
  styleUrls: ['./active-directory-view.component.scss']
})
export class AdViewComponent implements OnInit, AfterViewChecked {
  object: string;
  view = 'detail';
  adInfo: ActiveDirectoryType;
  treeWidth = '290px';
  detailWidth: string;
  pageWidth = window.innerWidth;
  treeHeight = window.innerHeight - 50;

  constructor(private router: Router,
              private activeDirectoryService: ActiveDirectoryService,
              private modalService: NgbModal,
              private treeObjectBehavior: TreeObjectBehavior,
              private cdr: ChangeDetectorRef) {
  }

  ngOnInit() {
    this.detailWidth = (this.pageWidth - 430) + 'px';
    this.treeObjectBehavior.$objectId.subscribe(id => {
      this.object = id;
      this.getInfo();
    });
  }

  ngAfterViewChecked(): void {
    this.cdr.detectChanges();
  }

  objectSelected($event: string) {
    this.object = $event;
  }

  getInfo() {
    const req = {
      'objectSid.equals': this.object,
      page: 1,
      size: 50
    };
    this.activeDirectoryService.query(req).subscribe(object => {
      if (object.body) {
        this.adInfo = object.body[0];
      }
    });
  }

  addToTracking() {
    const modalAddTracking = this.modalService.open(AdTrackerCreateComponent, {centered: true});
    modalAddTracking.componentInstance.targetTracking = [{
      name: this.adInfo.cn,
      id: this.adInfo.objectSid,
      type: resolveType(this.adInfo.objectClass),
      isAdmin: this.adInfo.adminCount !== null,
      objectSid: this.adInfo.objectSid
    }];
  }

  downloadReport() {
    const modalReport = this.modalService.open(AdReportCreateComponent, {centered: true});
    modalReport.componentInstance.data = [{
      name: this.adInfo.sAMAccountName,
      id: this.adInfo.objectSid,
      type: resolveType(this.adInfo.objectClass),
      isAdmin: this.adInfo.adminCount !== null,
      objectSid: this.adInfo.objectSid
    }];
  }

  onDragStart($event: DragEvent) {
  }

  onDragEnd($event: DragEvent) {
    if ($event.screenX >= 200 && $event.screenX <= 455) {
      this.detailWidth = (this.pageWidth - $event.screenX - 430) + 'px';
      this.treeWidth = $event.screenX + 'px';
    }
  }

  onDragging($event: DragEvent) {
    if ($event.screenX >= 200 && $event.screenX <= 455) {
      this.detailWidth = (this.pageWidth - $event.screenX - 340) + 'px';
      this.treeWidth = $event.screenX + 'px';
    }
  }

  onResize($event: ResizeEvent) {
    if ($event.rectangle.width >= 250) {
      this.detailWidth = (this.pageWidth - $event.rectangle.width - 340) + 'px';
      this.treeWidth = $event.rectangle.width + 'px';
    }
  }
}
