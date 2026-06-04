import {Component, Input, OnInit} from '@angular/core';
import {ActiveDirectoryLocalGroupType} from '../../../../../../types/active-directory-local-group.type';

@Component({
  selector: 'app-ad-computer-local-group',
  templateUrl: './ad-computer-local-group.component.html',
  styleUrls: ['./ad-computer-local-group.component.scss']
})
export class AdComputerLocalGroupComponent implements OnInit {
  @Input() localGroups: ActiveDirectoryLocalGroupType[];
  isCollapsed = -1;

  constructor() {
  }

  ngOnInit() {
  }

  toggleViewMembers(length: number, index: number) {
    if (length > 0) {
      this.isCollapsed = this.isCollapsed === index ? -1 : index;
    }
  }

  getMembersOfGroup(members: { name?: string; objectClass?: string }[]): string[] {
    const member: string[] = [];
    members.forEach(mem => member.push(mem.name));
    return member;
  }
}
