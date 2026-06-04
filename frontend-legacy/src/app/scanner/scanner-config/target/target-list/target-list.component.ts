import {HttpResponse} from '@angular/common/http';
import {Component, OnInit} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {JhiParseLinks} from 'ng-jhipster';
import {ITEMS_PER_PAGE} from '../../../../shared/constants/pagination.constants';
import {SortEvent} from '../../../../shared/directives/sortable/type/sort-event';
import {SortByType} from '../../../../shared/types/sort-by.type';
import {UsedByComponent} from '../../../shared/components/used-by/used-by.component';
import {TargetModel} from '../../../shared/model/target.model';
import {ScheduleService} from '../../schedule/shared/services/schedule.service';
import {TargetService} from '../shared/services/target.service';
import {TargetCreateComponent} from '../target-create/target-create.component';
import {TargetDeleteComponent} from '../target-delete/target-delete.component';

@Component({
  selector: 'app-target-list',
  templateUrl: './target-list.component.html',
  styleUrls: ['./target-list.component.scss']
})
export class TargetListComponent implements OnInit {
  fields: SortByType[] = [
    {
      fieldName: 'Target name',
      field: 'name'
    },
    {
      fieldName: 'Created at',
      field: 'created'
    },
    {
      fieldName: 'Port list',
      field: 'port_list '
    },
    {
      fieldName: 'Hosts',
      field: 'hosts'
    }
  ];
  targets: TargetModel[] = [];
  totalItems: any;
  page: any;
  itemsPerPage = ITEMS_PER_PAGE;
  loading = false;
  routeData: any;
  links: any;
  predicate: any;
  previousPage: any;
  reverse: any;
  targetDetail: TargetModel;
  private sortBy: string;
  private requestParams: any;

  constructor(private modalService: NgbModal,
              private activatedRoute: ActivatedRoute,
              private router: Router,
              private parseLinks: JhiParseLinks,
              private scheduleService: ScheduleService,
              private targetService: TargetService) {
  }

  ngOnInit() {
    this.routeData = this.activatedRoute.data.subscribe(data => {
      this.page = data.pagingParams.page;
      this.previousPage = data.pagingParams.page;
      this.reverse = data.pagingParams.ascending;
      this.predicate = data.pagingParams.predicate;
      // this.sortBy = this.predicate + ',' + (this.reverse ? 'asc' : 'desc');
    });
    this.loadTargets();
  }


  loadTargets() {
    this.requestParams = {
      page: this.page - 1,
      size: this.itemsPerPage,
      sort: this.sortBy,
      targets: true,
      details: true
    };
    this.getTargets();
  }

  getTargets() {
    this.loading = true;
    this.targetService.query(this.requestParams).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  newTarget() {
    const modal = this.modalService.open(TargetCreateComponent, {centered: true});
    modal.componentInstance.targetCreated.subscribe(created => {
      this.getTargets();
    });
  }

  onSortBy($event: SortEvent) {
    this.sortBy = $event.column + ',' + $event.direction;
    this.requestParams.sort = $event.column + ',' + $event.direction;
    this.loadTargets();
  }

  loadPage(page: number) {
    this.page = page;
    if (page !== this.previousPage) {
      this.previousPage = page;
      this.loadTargets();
    }
  }

  deleteTarget(target: TargetModel) {
    if (target.inUse === '0') {
      const modal = this.modalService.open(TargetDeleteComponent, {centered: true});
      modal.componentInstance.target = target;
      modal.componentInstance.targetDeleted.subscribe(created => {
        this.loadTargets();
      });
    }
  }

  editTarget(target) {
    const modal = this.modalService.open(TargetCreateComponent, {centered: true});
    modal.componentInstance.target = target;
    modal.componentInstance.targetCreated.subscribe(created => {
      this.loadTargets();
    });
  }

  showUse(target: any) {
    const modal = this.modalService.open(UsedByComponent, {centered: true});
    modal.componentInstance.using = target.tasks;
    modal.componentInstance.dependency = 'Tasks';
    modal.componentInstance.type = 'target';
    modal.componentInstance.name = target.name;
  }

  // toggleShowDetail(target: TargetModel) {
  //   if (target.uuid === this.targetDetail) {
  //     this.targetDetail = '';
  //   } else {
  //     this.targetDetail = target.uuid;
  //   }
  // }

  resolveSshCredential(target: TargetModel) {
    let credential = '';
    if (target.sshCredential.name !== null) {
      credential += target.sshCredential.name;
      if (target.sshCredential.port !== null) {
        credential += ' on port ' + target.sshCredential.port;
      }
    } else {
      credential = '';
    }
    return credential;
  }

  onTargetFilterChange($event: any) {
    Object.keys($event).forEach(key => {
      if ($event[key] !== '' && $event[key] !== null) {
        this.requestParams[key] = $event[key];
      } else {
        this.requestParams[key] = undefined;
      }
    });
    this.getTargets();
  }

  private onSuccess(data, headers) {
    this.links = this.parseLinks.parse(headers.get('link'));
    this.totalItems = headers.get('X-Total-Count');
    this.targets = data;
    this.loading = false;
  }

  private onError(error) {
    // this.alertService.error(error.error, error.message, null);
  }
}
