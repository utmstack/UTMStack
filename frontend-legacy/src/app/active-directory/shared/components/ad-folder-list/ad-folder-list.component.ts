import {Component, Input, OnInit} from '@angular/core';
import {ActiveDirectoryFolderAcl} from '../../types/active-directory-folder-acl';

@Component({
  selector: 'app-ad-folder-list',
  templateUrl: './ad-folder-list.component.html',
  styleUrls: ['./ad-folder-list.component.scss']
})
export class AdFolderListComponent implements OnInit {
  @Input() folders: ActiveDirectoryFolderAcl[] = [];
  foldersList: ActiveDirectoryFolderAcl[] = [];
  totalItems: number;
  page = 1;
  itemsPerPage = 5;
  pageStart = 0;
  pageEnd = 5;
  searchingFolder = false;
  searchFor: string;
  viewDetail: string;
  loading = false;

  constructor() {
  }

  ngOnInit() {
    this.foldersList = new FolderAccess(this.folders).folderAccess;
    this.totalItems = this.foldersList.length;
  }

  searchFolder($event) {
    this.searchingFolder = true;
    this.searchFor = $event;
    if (!$event) {
      this.totalItems = this.folders.length;
      this.searchingFolder = false;
      this.foldersList = new FolderAccess(this.folders).folderAccess;
    } else {
      const filtered = this.folders.filter(value => value.folder.toLocaleLowerCase().includes($event.toLocaleLowerCase()));
      this.foldersList = new FolderAccess(filtered).folderAccess;
      this.totalItems = this.foldersList.length;
      this.searchingFolder = false;
    }
  }

  loadPage(page: number) {
    this.pageEnd = page * this.itemsPerPage;
    this.pageStart = this.pageEnd - this.itemsPerPage;
  }

}

export class FolderAccess {
  constructor(folderAccess: ActiveDirectoryFolderAcl[]) {
    this._folderAccess = folderAccess;
  }

  // tslint:disable-next-line:variable-name
  public _folderAccess: ActiveDirectoryFolderAcl[] = [];

  get folderAccess(): ActiveDirectoryFolderAcl[] {
    return this._folderAccess;
  }

}
