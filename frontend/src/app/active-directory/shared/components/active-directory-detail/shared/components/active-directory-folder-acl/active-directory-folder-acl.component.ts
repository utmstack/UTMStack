import {Component, Input, OnInit} from '@angular/core';
import {ActiveDirectoryFolderAcl} from '../../../../../types/active-directory-folder-acl';
import {ActiveDirectoryUserComputer} from '../../../../../types/active-directory-user-computer';

@Component({
  selector: 'app-active-directory-folder-acl',
  templateUrl: './active-directory-folder-acl.component.html',
  styleUrls: ['./active-directory-folder-acl.component.scss']
})
export class ActiveDirectoryFolderAclComponent implements OnInit {
  @Input() computer: ActiveDirectoryUserComputer;
  folders: ActiveDirectoryFolderAcl[];

  constructor() {
  }

  ngOnInit() {
    this.folders = this.computer.localFolders.filter(value => value.access.length > 0);
  }

}

