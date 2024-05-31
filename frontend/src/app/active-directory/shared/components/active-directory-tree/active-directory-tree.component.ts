import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {AdReportCreateComponent} from '../../../reports/ad-report-create/ad-report-create.component';
import {AdTrackerCreateComponent} from '../../../tracker/ad-tracker-create/ad-tracker-create.component';
import {TreeObjectBehavior} from '../../behavior/tree-object.behvior';
import {ACTIVE_DIRECTORY_SIZE} from '../../const/active-directory-index-const';
import {getTreeIcon, resolveType} from '../../functions/ad-util.function';
import {ActiveDirectoryService} from '../../services/active-directory.service';
import {ActiveDirectoryTreeType} from '../../types/active-directory-tree.type';
import {ActiveDirectoryUsers} from '../../types/active-directory-users';
import {arrayToTree, TreeItem} from './shared/functions/array-to-tree.function';


// import * as Tree from
@Component({
  selector: 'app-active-directory-tree',
  templateUrl: './active-directory-tree.component.html',
  styleUrls: ['./active-directory-tree.component.scss']
})
export class AdTreeComponent implements OnInit {
  tree: TreeItem[] = [];
  treeOriginal: TreeItem[] = [];
  check = false;
  marked: ActiveDirectoryTreeType[] = [];
  deployed: string[] = [];
  deployedBeforeSearch: string[] = [];
  @Output() selected = new EventEmitter<string>();
  itemView = '';
  searching = false;
  loading = true;
  noDataFound: boolean;
  search: string;
  users: ActiveDirectoryUsers[] = [];

  constructor(private activeDirectoryService: ActiveDirectoryService,
              private modalService: NgbModal,
              private treeObjectBehavior: TreeObjectBehavior) {
  }

  ngOnInit() {
    this.getAllInfo();
  }

getAllInfo() {
    const req = {
      sourceId: 1,
      page: 0,
      size: ACTIVE_DIRECTORY_SIZE,
      sort: 'name,asc'
    };
    this.activeDirectoryService.query(req).subscribe(data => {
      this.searching = false;
      this.loading = false;
      if (data.body) {
        this.noDataFound = false;
        this.buildTree(data.body).then(temArr => {
          this.tree = arrayToTree(temArr,
            {parentId: 'parentId', id: 'id', dataField: null});
          this.tree = this.tree.filter( t => t.children.length > 0);
          this.deploy(this.tree[0]);
        });
      } else {
        this.noDataFound = true;
      }
    }, error => {
        this.loading = false;
        this.noDataFound = false;
        this.users = [];
    });
  }

  buildTree(activeDirectory: ActiveDirectoryUsers[]): Promise<ActiveDirectoryTreeType[]> {
    return new Promise<ActiveDirectoryTreeType[]>(resolve => {
      activeDirectory.unshift({
        id: 'Users',
        sid: null,
        createdDate: null,
        modifiedDate: null,
        source: null,
        name: 'Users'
      });
      activeDirectory.unshift({
        id: 'Workstations',
        sid: null,
        createdDate: null,
        modifiedDate: null,
        source: null,
        name: 'Workstations'
      });
      const arr: ActiveDirectoryTreeType[] = activeDirectory.reduce((group: any, value: any, currentIndex) => {
        const name = value.name;
        const existingGroup = group.find((group: any) => (name.startsWith('WS') && group.name === 'Workstations') ||  (group.name === 'Users'));
        if (group.length  > 0 && existingGroup) {
          group.push({
            parentId: name.startsWith('WS') ? 'Workstations-0' : 'Users-1',
            name,
            objectSid: value.sid,
            id: name + '-' + currentIndex,
            type: name.startsWith('WS') ? 'COMPUTER' : 'USER',
            indexPattern: value.source.indexPattern
          });
        } else {
          group.push({
            id: name + '-' + currentIndex,
            parentId: null,
            name,
            type: 'GROUP'
          });
        }

        return group;
      }, []);
      resolve(arr);
    });
  }

  toggleCheck() {
    this.check = this.check ? false : true;
    if (!this.check) {
      this.marked = [];
    }
  }

  addToSelect(item: ActiveDirectoryTreeType) {
    const index = this.marked.findIndex(value => value.name === item.name);
    if (item.children.length > 0 && index !== -1) {
      this.marked.splice(index, 1);
      this.unMarkChild(item);
    } else if (item.children.length === 0 && index !== -1) {
      this.marked.splice(index, 1);
    } else {
      this.marked.push(item);
      this.markChild(item);
    }
  }

  unMarkChild(item: ActiveDirectoryTreeType) {
    for (const it of item.children) {
      const indexChild = this.marked.findIndex(value => value.name === it.name);
      this.marked.splice(indexChild, 1);
      if (it.children.length > 0) {
        this.unMarkChild(it);
      }
    }
  }

  markChild(item: ActiveDirectoryTreeType) {
    for (const it of item.children) {
      this.marked.push(it);
      if (it.children.length > 0) {
        this.markChild(it);
      }
    }
  }

  deploy(item: TreeItem) {
    console.log(item);
    const index = this.deployed.findIndex(value => value === item.id);
    if (index === -1) {
      this.deployed.push(item.id);
    } else {
      this.deployed.splice(index, 1);
    }
    this.deployedBeforeSearch = this.deployed;
  }

  /**
   * Check tree object if user or group to trigger behavior;
   * @param item Receive select node in tree
   */
  select(item: ActiveDirectoryTreeType) {
    // (item.type === 'USER' || item.type === 'GROUP' || item.type === 'GROUP') &&
    if (item.children.length === 0) {
      this.itemView = item.id;
      this.selected.emit(item.objectSid);
      this.treeObjectBehavior.changeUser(item);
    }
  }


  resolveTreeIcon(item: ActiveDirectoryTreeType): string {
    return getTreeIcon(item);
  }

  isSelected(item: ActiveDirectoryTreeType): boolean {
    return this.marked.findIndex(value => value.name === item.name) !== -1;
  }

  addToTracking() {
    const modalAddTracking = this.modalService.open(AdTrackerCreateComponent, {centered: true});
    modalAddTracking.componentInstance.targetTracking = this.marked.filter(value => {
      return value.children.length === 0;
    });
  }

  downloadReport() {
    const modalReport = this.modalService.open(AdReportCreateComponent, {centered: true});
    modalReport.componentInstance.data = this.marked.filter(value => {
      return value.children.length === 0;
    });
  }

  searchInTree($event: any) {
    this.searching = true;
    setTimeout(() => {
      if ($event) {
        this.search = $event;
        this.filterAdByCn($event);
      } else {
        this.search = null;
        this.getAllInfo();
        this.deployed = [];
      }
    }, 1000);
  }

  filterAdByCn(cn: string) {
    const data = this.filterByName(cn);
    if (data.length > 0) {
        this.noDataFound = false;
        this.searching = false;
        this.tree = data;
        for (const node of this.tree) {
            this.deployAll(node);
          }
      } else {
        this.searching = false;
        this.noDataFound = true;
      }
  }

  // deploy all children in tree when search
  deployAll(item: TreeItem) {
    this.deployed.push(item.id);
    for (const it of item.children) {
      const index = this.deployed.findIndex(value => value === it.id);
      if (index === -1) {
        this.deployed.push(it.id);
      }
      if (it.children.length > 0) {
        this.deployAll(it);
      }
    }
  }

   filterByName(partialName: string) {
    return this.tree.filter(item => {
      if ( item.name.toLowerCase().includes(partialName.toLowerCase()) ||
          (item.children && item.children.some(child => child.name.toLowerCase().includes(partialName.toLowerCase())))
      ) {
        return true;
      }
      return false;
    });
  }


  findPath(node: TreeItem, nodeName): TreeItem[] {
    // If current node matches search node, return tail of path result
    if (node.name.includes(nodeName) && node.children.length === 0) {
      return [node];
    } else {

      // If current node not search node match, examine children. For first
      // child that returns an array (path), prepend current node to that
      // path result
      for (const child of node.children) {
        const childPath = this.findPath(child, nodeName);
        if (Array.isArray(childPath)) {
          childPath.unshift(child);
          return childPath;
        }
      }
    }
  }
}


