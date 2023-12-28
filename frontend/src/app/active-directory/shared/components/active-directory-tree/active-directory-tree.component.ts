import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {AdReportCreateComponent} from '../../../reports/ad-report-create/ad-report-create.component';
import {AdTrackerCreateComponent} from '../../../tracker/ad-tracker-create/ad-tracker-create.component';
import {TreeObjectBehavior} from '../../behavior/tree-object.behvior';
import {ACTIVE_DIRECTORY_SIZE} from '../../const/active-directory-index-const';
import {getTreeIcon, resolveType} from '../../functions/ad-util.function';
import {ActiveDirectoryService} from '../../services/active-directory.service';
import {ActiveDirectoryTreeType} from '../../types/active-directory-tree.type';
import {ActiveDirectoryType} from '../../types/active-directory.type';
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
  itemView: string;
  searching = false;
  noDataFound: boolean;
  search: string;

  constructor(private activeDirectoryService: ActiveDirectoryService,
              private modalService: NgbModal,
              private treeObjectBehavior: TreeObjectBehavior) {
  }

  ngOnInit() {
    // reset objectId
    this.treeObjectBehavior.$objectId.next(null);
    this.getAllInfo();
  }

  getAllInfo() {
    const req = {
      page: 1,
      size: ACTIVE_DIRECTORY_SIZE,
      // 'objectClass.specified': true,
    };
    this.activeDirectoryService.query(req).subscribe(data => {
      this.searching = false;
      if (data.body) {
        this.noDataFound = false;
        this.buildTree(data.body).then(temArr => {
          this.tree = arrayToTree(temArr,
            {parentId: 'parentId', id: 'id', dataField: null});
        });
      } else {
        this.noDataFound = true;
      }
    });
  }

  buildTree(activeDirectory: ActiveDirectoryType[]): Promise<ActiveDirectoryTreeType[]> {
    return new Promise<ActiveDirectoryTreeType[]>(resolve => {
      const arr: ActiveDirectoryTreeType[] = [];
      for (const ad of activeDirectory) {
        // tslint:disable-next-line:variable-name
        if (Object(ad).hasOwnProperty('distinguishedName') && ad.distinguishedName) {
          const path = ad.distinguishedName.split(',').reverse();
          // tslint:disable-next-line:prefer-for-of
          for (let i = 0; i < path.length; i++) {
            const nodeName = path[i].substring(3, path[i].length);
            const parentName = path[i - 1] ? path[i - 1]
              .substring(3, path[i - 1].length) + '-' + (i - 1) : null;
            const node = {
              parentId: parentName,
              name: nodeName,
              objectSid: ad.objectSid,
              id: nodeName + '-' + i,
              type: resolveType(ad.objectClass),
              isAdmin: ad.adminCount !== null,
              // children: []
            };
            if (arr.findIndex(value => value.parentId === node.parentId && value.id === node.id) === -1) {
              arr.push(node);
            }
          }
        }
      }
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

  deploy(item: ActiveDirectoryTreeType) {
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
      this.treeObjectBehavior.$objectId.next(item.objectSid);
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
    const req = {
      page: 1,
      size: ACTIVE_DIRECTORY_SIZE,
      // 'cn.contains': cn,
      'displayName.contains': cn
    };
    this.activeDirectoryService.query(req).subscribe(data => {
      if (data.body) {
        this.noDataFound = false;
        this.buildTree(data.body).then(temArr => {
          this.searching = false;
          this.tree = arrayToTree(temArr,
            {parentId: 'parentId', id: 'id', dataField: null});
          for (const node of this.tree) {
            this.deployAll(node);
          }
        });
      } else {
        this.searching = false;
        this.noDataFound = true;
      }
    });
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


