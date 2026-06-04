import {HttpResponse} from '@angular/common/http';
import {Component, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {ITEMS_PER_PAGE} from '../../../shared/constants/pagination.constants';
import {UtmDateFormatEnum} from '../../../shared/enums/utm-date-format.enum';
import {SortByType} from '../../../shared/types/sort-by.type';
import {TimeFilterType} from '../../../shared/types/time-filter.type';
import {resolveRangeByTime} from '../../../shared/util/resolve-date';
import {AdReportCreateComponent} from '../../reports/ad-report-create/ad-report-create.component';
import {resolveType} from '../../shared/functions/ad-util.function';
import {ActiveDirectoryService} from '../../shared/services/active-directory.service';
import {ActiveDirectoryTreeType} from '../../shared/types/active-directory-tree.type';
import {ActiveDirectoryType} from '../../shared/types/active-directory.type';
import {AdTrackerCreateComponent} from '../../tracker/ad-tracker-create/ad-tracker-create.component';
import {AdTrackerType} from '../../tracker/shared/type/ad-tracker.type';

@Component({
  selector: 'app-ad-user-detail',
  templateUrl: './ad-user-detail.component.html',
  styleUrls: ['./ad-user-detail.component.scss']
})
export class AdUserDetailComponent implements OnInit {
  adInfos: ActiveDirectoryType[] = [];
  fields: SortByType[] = [
    {
      fieldName: 'Name',
      field: 'cn.keyword'
    },
    {
      fieldName: 'Account',
      field: 'sAMAccountName.keyword'
    },
    {
      fieldName: 'Creation time',
      field: 'whenCreated'
    },
    {
      fieldName: 'Last logon',
      field: 'realLastLogon'
    }
  ];
  loading = true;
  totalItems: any;
  // init on 1
  page = 1;
  previousPage: number;
  routeData: any;
  links: any;
  predicate: any;
  reverse: any;
  itemsPerPage = ITEMS_PER_PAGE;
  search: any;
  sortBy: string;
  selected: ActiveDirectoryType[] = [];
  allSelected = false;
  viewDetail = '';
  tracked: ActiveDirectoryType[] = [];
  viewEvent: ActiveDirectoryType;
  apiUrl: string;
  daysToBeInactiveUser: number;
  requestParams: {
    page: number,
    size: number,
    sort: string,
    daysToBeInactiveUser?: string,
    isAdmin?: boolean,
    from: string,
    to: string
  } = {
    page: this.page,
    size: this.itemsPerPage,
    sort: '',
    daysToBeInactiveUser: '',
    isAdmin: undefined,
    from: '',
    to: ''
  };
  status: string;
  title = 'Detail of ';
  timeFilter: TimeFilterType = resolveRangeByTime('week');
  sAMAccountName: string;
  events: string[];
  formatDateEnum = UtmDateFormatEnum;

  constructor(private modalService: NgbModal,
              private activeDirectoryService: ActiveDirectoryService,
              private activeRoute: ActivatedRoute) {
  }

  ngOnInit() {
    // build dynamic url depending on user status
    this.activeRoute.queryParams.subscribe(params => {
      if (params) {
        if (params.status) {
          switch (params.status) {
            case 'inactive':
              this.apiUrl = 'api/ad/dashboard/amount-of-inactive-users-details';
              this.daysToBeInactiveUser = params.inactiveDays;
              this.requestParams.daysToBeInactiveUser = params.inactiveDays;
              this.title = 'users with more than ' + params.inactiveDays + ' days inactive';
              break;
            case 'lockdown':
              this.apiUrl = 'api/ad/dashboard/amount-of-users-lockout-details';
              this.title += 'users with status lockdown';
              break;
            case 'disabled':
              this.apiUrl = 'api/ad/dashboard/amount-of-users-disabled-details';
              this.title += 'users with status disabled';
              break;
          }
          this.loadUsers();
        }
        if (params.userType) {
          this.apiUrl = 'api/ad/dashboard/amount-of-admins-vs-users-details';
          this.requestParams.isAdmin = params.userType !== 'users';
          this.title += params.userType !== 'users' ? 'users with administration permissions' : 'users with standard permissions';
          this.loadUsers();
        }
        if (params.permissions) {
          this.apiUrl = 'api/ad/dashboard/amount-of-users-scaled-privileges-details';
          this.requestParams.from = params.from;
          this.requestParams.to = params.to;
          this.timeFilter.timeFrom = params.from;
          this.timeFilter.timeTo = params.to;
          this.timeFilter.range = params.rangeDate;
          this.events = ['4728', '4732'];
          this.title = 'Objects that scaled permissions';
          this.loadUsers();
        }
        if (params.userChange) {
          this.apiUrl = 'api/ad/active-directory-info-by-filter';
          this.sAMAccountName = params.userChange;
          this.title = 'Detail of user ' + params.userChange;
        }
      }
    });
  }

  loadUsers() {
    this.requestParams.page = this.page;
    this.requestParams.size = this.itemsPerPage;
    this.requestParams.sort = this.sortBy;
    this.getUserList();
  }

  onSortBy($event) {
    this.requestParams.sort = $event;
    this.getUserList();
  }

  loadPage(page: any) {
    this.requestParams.page = page;
    this.getUserList();
  }

  getUserList() {
    this.activeDirectoryService.queryUser(this.requestParams, this.apiUrl).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  objectIconResolver(ad: ActiveDirectoryType): string {
    switch (this.getObjectType(ad.objectClass)) {
      case 'COMPUTER':
        return 'icon-display';
      case 'GROUP':
        return 'icon-users2';
      default:
        return ad.adminCount !== null ? 'icon-user-tie' : 'icon-user';
    }
  }

  downloadReport() {
    const modalReport = this.modalService.open(AdReportCreateComponent, {centered: true});
    modalReport.componentInstance.data = this.extractUserToReport();
  }

  extractUserToReport(): AdTrackerType[] {
    const arr: ActiveDirectoryTreeType[] = [];
    for (const ad of this.selected) {
      arr.push({
        name: ad.cn,
        id: ad.objectSid,
        type: this.getObjectType(ad.objectClass),
        objectSid: ad.objectSid
      });
    }
    return arr;
  }

  downloadSingleReport(ad: ActiveDirectoryType) {
    const modalReport = this.modalService.open(AdReportCreateComponent, {centered: true});
    modalReport.componentInstance.data = [{
      name: ad.cn,
      id: ad.objectSid,
      type: resolveType(ad.objectClass),
      objectSid: ad.objectSid
    }];
  }

  isSelected(tracker: AdTrackerType): boolean {
    return this.selected.findIndex(value => value.id === tracker.id) !== -1;
  }

  addToSelected(tracker: ActiveDirectoryType) {
    const index = this.selected.findIndex(value => value.id === tracker.id);
    if (index === -1) {
      this.selected.push(tracker);
    } else {
      this.selected.splice(index, 1);
    }
  }

  getObjectType(adInfo): string {
    return resolveType(adInfo);
  }

  adToTracking() {
    const modalAddTracking = this.modalService.open(AdTrackerCreateComponent, {centered: true});
    modalAddTracking.componentInstance.targetTracking = this.convertSelectedToTreeType();
    modalAddTracking.componentInstance.trackingCreated.subscribe(() => {
      this.selected.forEach(value => this.tracked.push(value));
    });
  }

  convertSelectedToTreeType(): ActiveDirectoryTreeType[] {
    const arr: ActiveDirectoryTreeType[] = [];
    for (const select of this.selected) {
      arr.push({
        objectSid: select.objectSid,
        name: select.cn,
        type: this.getObjectType(select.objectClass)
      });
    }
    return arr;
  }

  toggleAllSelection() {
    this.allSelected = this.allSelected ? false : true;
    if (this.allSelected) {
      this.selected = this.adInfos;
    } else {
      this.selected = [];
    }
  }

  trackUser(adInfo: ActiveDirectoryType) {
    if (!this.isTracked(adInfo)) {
      const modalAddTracking = this.modalService.open(AdTrackerCreateComponent, {centered: true});
      modalAddTracking.componentInstance.targetTracking = [
        {
          objectSid: adInfo.objectSid,
          name: adInfo.cn,
          type: this.getObjectType(adInfo.objectClass)
        }
      ];
      modalAddTracking.componentInstance.trackingCreated.subscribe(() => {
        this.tracked.push(adInfo);
      });
    }
  }

  isTracked(adInfo: ActiveDirectoryType) {
    return this.tracked.findIndex(value => value.objectSid === adInfo.objectSid) !== -1;
  }

  onAdFilterChange($event: any) {
    Object.keys($event).forEach(key => {
      if ($event[key] !== '' && $event[key] !== null) {
        this.requestParams[key] = $event[key];
      } else {
        this.requestParams[key] = undefined;
      }
    });
    this.getUserList();
  }

  private onSuccess(data, headers) {
    this.totalItems = headers.get('X-Total-Count');
    this.adInfos = data;
    this.loading = false;
  }

  private onError(error) {
    this.loading = false;
    // this.alertService.error(error.error, error.message, null);
  }

}
