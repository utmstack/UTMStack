import {Component, Input, OnInit} from '@angular/core';
import {UtmDateFormatEnum} from '../../../../shared/enums/utm-date-format.enum';
import {resolveMembersOf, resolveType} from '../../functions/ad-util.function';
import {ActiveDirectoryLocalGroupType} from '../../types/active-directory-local-group.type';
import {ActiveDirectoryType} from '../../types/active-directory.type';

@Component({
  selector: 'app-active-directory-detail',
  templateUrl: './active-directory-detail.component.html',
  styleUrls: ['./active-directory-detail.component.scss']
})
export class AdDetailComponent implements OnInit {
  @Input() adInfo: any;
  @Input() user: any;
  formatDateEnum = UtmDateFormatEnum;
  totalItems: number;
  page = 1;
  itemsPerPage = 10;
  pageStart = 0;
  pageEnd = 10;
  viewAcl: any;
  searching = false;
  viewComputer: any;
  computerDetail: ActiveDirectoryLocalGroupType;
  localGroups: ActiveDirectoryLocalGroupType[];
  computerSearch: string;

  constructor() {
  }

  ngOnInit() {
    setTimeout(() => {
      this.localGroups = this.adInfo.localGroups;
    }, 1000);
  }

  getObjectType(): string {
    return resolveType(this.adInfo.objectClass);
  }

  resolvePath(): string {
    let path = '';
    const way = this.adInfo.distinguishedName.split(',').reverse();
    for (const w of way) {
      path += w.substring(3, w.length) + '/';
    }
    return path;
  }

  resolveObjectMembers() {
    return resolveMembersOf(this.adInfo.member);
  }

  resolveObjectMemberOff() {
    return resolveMembersOf(this.adInfo.memberOf);
  }

  searchComputer($event) {
    this.computerSearch = $event;
    if ($event) {
      this.searching = false;
      this.localGroups = this.adInfo.localGroups.filter(group => {
        return group.name.toLocaleLowerCase().includes($event.toLocaleLowerCase() ||
          group.description.toLocaleLowerCase().includes($event.toLocaleLowerCase()));
      });
    } else {
      this.searching = false;
      this.localGroups = this.adInfo.localGroups;
    }
  }

  viewComputerDetail(computer: ActiveDirectoryLocalGroupType) {
    this.computerDetail = computer;
    this.viewComputer = true;
  }

  public resolveSoIcon(so: string): string {
    if (so.toLowerCase().includes('windows')) {
      return 'icon-windows8';
    } else if (so.toLowerCase().includes('linux')) {
      return 'icon-tux';
    } else if (so.toLowerCase().includes('mac')) {
      return 'icon-apple2';
    } else if (so.toLowerCase().includes('possible conflict')) {
      return 'icon-power2';
    } else {
      return 'icon-question3';
    }
  }
}
