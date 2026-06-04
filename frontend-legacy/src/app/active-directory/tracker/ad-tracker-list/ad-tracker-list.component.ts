import {HttpResponse} from '@angular/common/http';
import {Component, OnInit} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {ITEMS_PER_PAGE} from '../../../shared/constants/pagination.constants';
import {UtmDateFormatEnum} from '../../../shared/enums/utm-date-format.enum';
import {SortByType} from '../../../shared/types/sort-by.type';
import {UtmWebfluxService} from '../../../shared/webflux/utm-webflux.service';
import {AdReportCreateComponent} from '../../reports/ad-report-create/ad-report-create.component';
import {ActiveDirectoryTreeType} from '../../shared/types/active-directory-tree.type';
import {AdTrackerDeleteComponent} from '../ad-tracker-delete/ad-tracker-delete.component';
import {AdTrackerService} from '../shared/services/ad-tracker.service';
import {AdTrackerType} from '../shared/type/ad-tracker.type';

@Component({
  selector: 'app-active-directory-tracker-list',
  templateUrl: './ad-tracker-list.component.html',
  styleUrls: ['./ad-tracker-list.component.scss']
})
export class AdTrackerListComponent implements OnInit {
  trackers: AdTrackerType[] = [];
  fields: SortByType[] = [
    {
      fieldName: 'Name',
      field: 'name'
    },
    {
      fieldName: 'Last modification',
      field: 'modificationTime'
    }
  ];
  loading = false;
  totalItems: any;
  page = 1;
  previousPage: number;
  routeData: any;
  links: any;
  predicate: any;
  reverse: any;
  itemsPerPage = ITEMS_PER_PAGE;
  search: any;
  sortBy: string;
  changeConfigNotify = false;
  selected: AdTrackerType[] = [];
  viewEvent: AdTrackerType;
  formatDateEnum = UtmDateFormatEnum;
  private requestParams: any;

  constructor(private adTrackerService: AdTrackerService,
              private modalService: NgbModal,
              private activatedRoute: ActivatedRoute,
              private router: Router) {
  }

  ngOnInit() {
    this.loadTrackers();
    this.getTrackerList();
   }

  loadTrackers() {
    this.loading = true;
    this.requestParams = {
      page: this.page - 1,
      size: this.itemsPerPage,
      sort: this.sortBy,
    };
    this.getTrackerList();
  }

  onSortBy($event) {
  }

  deleteTracker(tracker) {
    const modal = this.modalService.open(AdTrackerDeleteComponent, {centered: true});
    modal.componentInstance.tracker = tracker;
    modal.componentInstance.trackerDeleted.subscribe(deleted => {
      this.getTrackerList();
    });
  }


  loadPage(page: any) {
    this.requestParams.page = page - 1;
    this.getTrackerList();
  }

  transition() {
    this.router.navigate(['/active-directory/tracker'], {
      queryParams: {
        page: this.page,
        size: this.itemsPerPage,
        sort: this.sortBy
      }
    });
    this.getTrackerList();
  }

  getTrackerList() {
    this.adTrackerService.query(this.requestParams).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  objectIconResolver(objectType: string): string {
    switch (objectType) {
      case 'COMPUTER':
        return 'icon-display';
      case 'GROUP':
        return 'icon-users2';
      default:
        return 'icon-user';
    }
  }

  onTrackerFilterChange($event: any) {
    Object.keys($event).forEach(key => {
      if ($event[key] !== '' && $event[key] !== null) {
        this.requestParams[key] = $event[key];
      } else {
        this.requestParams[key] = undefined;
      }
    });
    this.getTrackerList();
  }

  downloadReport() {
    const modalReport = this.modalService.open(AdReportCreateComponent, {centered: true});
    modalReport.componentInstance.data = this.extractTrackerToReport();
  }

  extractTrackerToReport(): AdTrackerType[] {
    const arr: ActiveDirectoryTreeType[] = [];
    for (const tracker of this.selected) {
      arr.push({
        name: tracker.objectName,
        id: tracker.objectId,
        type: tracker.objectType,
        objectSid: tracker.objectId
      });
    }
    return arr;
  }

  downloadSingleReport(tracker: AdTrackerType) {
    const modalReport = this.modalService.open(AdReportCreateComponent, {centered: true});
    modalReport.componentInstance.data = [{
      name: tracker.objectName,
      id: tracker.objectId,
      type: tracker.objectType,
      objectSid: tracker.objectId
    }];
  }

  isSelected(tracker: AdTrackerType): boolean {
    return this.selected.findIndex(value => value.id === tracker.id) !== -1;
  }

  addToTracker(tracker: AdTrackerType) {
    const index = this.selected.findIndex(value => value.id === tracker.id);
    if (index === -1) {
      this.selected.push(tracker);
    } else {
      this.selected.splice(index, 1);
    }
  }

  private onSuccess(data, headers) {
    this.totalItems = headers.get('X-Total-Count');
    this.trackers = data;
    this.loading = false;
  }

  private onError(error) {
    this.loading = false;
    // this.alertService.error(error.error, error.message, null);
  }
}
