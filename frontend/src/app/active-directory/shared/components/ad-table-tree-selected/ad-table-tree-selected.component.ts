import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {AdTrackerType} from '../../../tracker/shared/type/ad-tracker.type';
import {getTreeIcon} from '../../functions/ad-util.function';
import {ActiveDirectoryTreeType} from '../../types/active-directory-tree.type';

@Component({
  selector: 'app-ad-table-tree-selected',
  templateUrl: './ad-table-tree-selected.component.html',
  styleUrls: ['./ad-table-tree-selected.component.scss']
})
export class AdTableTreeSelectedComponent implements OnInit {
  @Input() rows: ActiveDirectoryTreeType[];
  @Input() response: 'string' | 'array';
  @Output() tableChange = new EventEmitter<AdTrackerType[] | string>();
  totalItems: number;
  page = 1;
  itemsPerPage = 5;
  pageStart = 0;
  pageEnd = 5;

  constructor() {
  }

  ngOnInit() {
    this.totalItems = this.rows.length;
    this.emitResponse();
  }

  resolveIcon(row: ActiveDirectoryTreeType): string {
    return getTreeIcon(row);
  }


  deleteTracker(index: number) {
    this.rows.splice(index, 1);
    this.emitResponse();
  }

  emitResponse() {
    if (this.response === 'string') {
      this.tableChange.emit(this.extractObjectSid());
    } else {
      this.tableChange.emit(this.extractTracker());
    }
  }

  extractObjectSid(): string {
    let objectSid = '';
    for (const tracker of this.rows) {
      objectSid += tracker.objectSid + ',';
    }
    return objectSid.substring(0, objectSid.length - 1);
  }

  extractTracker(): AdTrackerType[] {
    const trackers: AdTrackerType[] = [];
    for (const tracker of this.rows) {
      trackers.push(
        {
          objectId: tracker.objectSid,
          objectName: tracker.name,
          objectType: tracker.type,
        });
    }
    return trackers;
  }

  loadPage(page: number) {
    this.pageEnd = page * this.itemsPerPage;
    this.pageStart = this.pageEnd - this.itemsPerPage;
  }
}
