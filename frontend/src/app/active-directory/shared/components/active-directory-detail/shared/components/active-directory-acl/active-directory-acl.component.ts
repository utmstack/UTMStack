import {Component, Input, OnInit} from '@angular/core';
import {UtmDateFormatEnum} from '../../../../../../../shared/enums/utm-date-format.enum';
import {ActiveDirectoryType} from '../../../../../types/active-directory.type';

@Component({
  selector: 'app-active-directory-acl',
  templateUrl: './active-directory-acl.component.html',
  styleUrls: ['./active-directory-acl.component.scss']
})
export class ActiveDirectoryAclComponent implements OnInit {
  @Input() adInfo: ActiveDirectoryType;
  formatDateEnum = UtmDateFormatEnum;
  totalItems: number;
  page = 1;
  itemsPerPage = 10;
  pageStart = 0;
  pageEnd = 10;

  constructor() {
  }

  ngOnInit() {
    this.totalItems = this.adInfo.userACLs.length;
  }

  loadPage(page: number) {
    this.pageEnd = this.page * this.itemsPerPage;
    this.pageStart = this.pageEnd - this.itemsPerPage;
  }


}
