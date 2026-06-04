import {Component, Input, OnInit} from '@angular/core';
import {ActiveDirectoryLocalUserType} from '../../../../../../types/active-directory-local-user.type';
import {ActiveDirectoryType} from '../../../../../../types/active-directory.type';

@Component({
  selector: 'app-ad-computer-local-users',
  templateUrl: './ad-computer-local-users.component.html',
  styleUrls: ['./ad-computer-local-users.component.scss']
})
export class AdComputerLocalUsersComponent implements OnInit {
  @Input() adInfo: ActiveDirectoryType;
  localUsers: ActiveDirectoryLocalUserType[];
  userSearch: string;

  constructor() {
  }

  ngOnInit() {
    this.localUsers = this.adInfo.localUsers;
  }

  onSearchUser($event: string) {
    this.userSearch = $event;
    if ($event) {
      this.localUsers = this.localUsers.filter(value => value.name.toLowerCase().includes($event.toLowerCase()));
    } else {
      this.localUsers = this.adInfo.localUsers;
    }
  }
}
