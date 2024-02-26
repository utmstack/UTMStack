import {AfterViewInit, Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {searchEventType} from '../../../../shared/constants/active-directory-event.const';
import {ITEMS_PER_PAGE} from '../../../../shared/constants/pagination.constants';
import {ElasticTimeEnum} from '../../../../shared/enums/elastic-time.enum';
import {UtmDateFormatEnum} from '../../../../shared/enums/utm-date-format.enum';
import {ElasticFilterCommonType} from '../../../../shared/types/filter/elastic-filter-common.type';
import {TimeFilterType} from '../../../../shared/types/time-filter.type';
import {TreeObjectBehavior} from '../../behavior/tree-object.behvior';
import {WinlogbeatService} from '../../services/winlogbeat.service';
import {ActiveDirectoryTreeType} from '../../types/active-directory-tree.type';
import {WinlogbeatEventType} from '../../types/winlogbeat-event.type';

@Component({
  selector: 'app-event-timeline',
  templateUrl: './event-timeline.component.html',
  styleUrls: ['./event-timeline.component.scss']
})
export class EventTimelineComponent implements OnInit, AfterViewInit {
  @Input() events: string[];
  @Input() time: TimeFilterType;
  @Output() eventChange = new EventEmitter<WinlogbeatEventType>();
  objectId: ActiveDirectoryTreeType;
  sevenDaysRange: ElasticFilterCommonType = {time: ElasticTimeEnum.DAY, last: 7, label: 'last 7 days'};
  items: WinlogbeatEventType[] = [];
  loadingMore = false;
  totalItems: any;
  page = 1;
  itemsPerPage = ITEMS_PER_PAGE;
  filterTime: TimeFilterType;
  loading = true;
  itemSelected: string;
  formatDateEnum = UtmDateFormatEnum;
  noMoreResult = false;

  constructor(private winlogbeatService: WinlogbeatService,
              private treeObjectBehavior: TreeObjectBehavior) {

  }

  ngOnInit(): void {
    this.treeObjectBehavior.userSelected().subscribe(user => {
      this.eventChange.emit(null);
      this.itemSelected = '';
      if (user) {
        this.objectId = user;
        this.items = [];
        this.page = 1;
        this.getEvents();
      }
    });
  }

  getEvents() {
    if (this.filterTime && this.objectId.objectSid) {
      const req = {
        page: this.page,
        size: this.itemsPerPage,
        sort: '@timestamp,desc',
        sid: this.objectId.objectSid,
        indexPattern: this.objectId.indexPattern,
        from: this.filterTime.timeFrom,
        to: this.filterTime.timeTo,
        'eventId.in': this.events ? this.events.toString() : undefined
      };
      this.winlogbeatService.query(req).subscribe(response => {
        this.loadingMore = false;
        this.loading = false;
        if (response.body === null || response.body.length === 0) {
          this.eventChange.emit(null);
        } else {
          this.items = response.body;
          this.totalItems = Number(response.headers.get('X-Total-Count'));
        }
      });
    } else {
      this.loading = false;
    }
  }

  ngAfterViewInit() {
    // new WOW().init();
  }

  onScroll() {
    this.loadingMore = true;
    this.page += 1;
    this.getEvents();
  }

  resolveClassByEventType(eventId: number): string {
    const event = searchEventType(eventId);
    if (event) {
      switch (event.type) {
        case 'COMPUTER':
          return 'utm_tmlabel_computer';
        case 'GROUP':
          return 'utm_tmlabel_group';
        case 'OBJECT':
          return 'utm_tmlabel_object';
        case 'USER':
          return 'utm_tmlabel_user';
      }
    } else {
      return 'utm_tmlabel_not_found';
    }
  }

  resolveClassSelected(eventId: number): string {
    const event = searchEventType(eventId);
    if (event) {
      switch (event.type) {
        case 'COMPUTER':
          return 'utm_tmlabel_computer_selected';
        case 'GROUP':
          return 'utm_tmlabel_group_selected';
        case 'OBJECT':
          return 'utm_tmlabel_object_selected';
        case 'USER':
          return 'utm_tmlabel_user_selected';
      }
    } else {
      return 'utm_tmlabel_not_found_selected';
    }
  }

  resolveMessage(eventId: number): string {
    const event = searchEventType(eventId);
    if (event) {
      return event.message;
    } else {
      return 'Event info was not found';
    }
  }


  selectEvent(item: WinlogbeatEventType) {
    this.itemSelected = this.getUniqueEventId(item);
    this.eventChange.emit(item);
  }

  getUniqueEventId(item: WinlogbeatEventType) {
    return item.id + '-' + item.logx.wineventlog.eventId + '-' + new Date(item.timestamp).getTime();
  }

  onFilterTimeChange($event: TimeFilterType) {
    this.filterTime = $event;
    this.page = 1;
    this.items = [];
    this.getEvents();
  }

  loadPage($event: number) {
    this.page = $event;
    this.getEvents();
  }

  onItemsPerPageChange($event: number) {
    this.itemsPerPage = $event;
    this.getEvents();

  }
}


