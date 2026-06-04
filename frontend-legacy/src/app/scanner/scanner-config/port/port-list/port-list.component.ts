import {HttpResponse} from '@angular/common/http';
import {Component, OnInit} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {JhiParseLinks} from 'ng-jhipster';
import {AccountService} from '../../../../core/auth/account.service';
import {ITEMS_PER_PAGE} from '../../../../shared/constants/pagination.constants';
import {SortEvent} from '../../../../shared/directives/sortable/type/sort-event';
import {SortByType} from '../../../../shared/types/sort-by.type';
import {UsedByComponent} from '../../../shared/components/used-by/used-by.component';
import {PortModel} from '../../../shared/model/port.model';

import {PortCreateComponent} from '../port-create/port-create.component';
import {PortDeleteComponent} from '../port-delete/port-delete.component';
import {PortRangeListComponent} from '../port-range/port-range-list/port-range-list.component';
import {PortService} from '../shared/services/port.service';


@Component({
  selector: 'app-port-list',
  templateUrl: './port-list.component.html',
  styleUrls: ['./port-list.component.scss']
})
export class PortListComponent implements OnInit {
  ports: PortModel[] = [];
  loading = false;
  totalItems: any;
  page: any;
  itemsPerPage = ITEMS_PER_PAGE;
  error: any;
  success: any;
  routeData: any;
  links: any;
  predicate: any;
  previousPage: any;
  reverse: any;
  search: string;
  request: any;
  fields: SortByType[] = [
    {
      fieldName: 'Port name',
      field: 'name'
    },
    {
      fieldName: 'Last modification',
      field: 'modificationTime'
    },
    {
      fieldName: 'TCP',
      field: 'insecureUse'
    },
    {
      fieldName: 'UDP',
      field: 'login'
    }
  ];
  private sortBy: string;

  constructor(private modalService: NgbModal,
              private activatedRoute: ActivatedRoute,
              private router: Router,
              private parseLinks: JhiParseLinks,
              private accountService: AccountService,
              private portService: PortService) {
  }

  ngOnInit() {
    this.routeData = this.activatedRoute.data.subscribe(data => {
      this.page = data.pagingParams.page;
      this.previousPage = data.pagingParams.page;
      this.reverse = data.pagingParams.ascending;
      this.predicate = data.pagingParams.predicate;
      this.sortBy = this.predicate + ',' + (this.reverse ? 'asc' : 'desc');
    });
    this.loadPorts();
  }

  loadPorts() {
    this.request = {
      page: this.page - 1,
      size: this.itemsPerPage,
      'name.notContains': 'nmap',
      sort: this.sortBy,
      targets: true,
      details: true
    };
    this.getPorts();
  }

  getPorts() {
    this.loading = true;
    this.portService.query(this.request).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  loadPage(page: number) {
    this.page = page;
    if (page !== this.previousPage) {
      this.previousPage = page;
      this.loadPorts();
    }
  }

  transition() {
    this.router.navigate(['/scanner/config/port'], {
      queryParams: {
        page: this.page,
        sort: this.sortBy
      }
    });
    this.loadPorts();
  }

  onSortBy($event: SortEvent) {
    this.sortBy = $event.column + ',' + $event.direction;
    this.request.sort = this.sortBy;
    this.loadPorts();
  }

  newPort() {
    const modal = this.modalService.open(PortCreateComponent, {centered: true});
    modal.componentInstance.portCreated.subscribe(created => {
      this.loadPorts();
    });
  }

  editPort(port: any) {
    const modal = this.modalService.open(PortRangeListComponent, {centered: true});
    modal.componentInstance.port = port;
    modal.componentInstance.portEdited.subscribe(() => {
      this.loadPorts();
    });
  }

  deletePort(port: any) {
    const modal = this.modalService.open(PortDeleteComponent, {centered: true});
    modal.componentInstance.port = port;
    modal.componentInstance.portDeleted.subscribe(() => {
      this.getPorts();
    });
  }

  showUse(port: PortModel) {
    const modal = this.modalService.open(UsedByComponent, {centered: true});
    modal.componentInstance.using = port.targets;
    modal.componentInstance.dependency = 'Targets';
    modal.componentInstance.type = 'port';
    modal.componentInstance.name = port.name;
  }

  onPortFilterChange($event: any) {
    Object.keys($event).forEach(key => {
      if ($event[key] !== '' && $event[key] !== null) {
        this.request[key] = $event[key];
      } else {
        this.request[key] = undefined;
      }
    });
    setTimeout(() => this.getPorts(), 2000);
  }

  cleanPortRelatedToOpenvas(ports: PortModel[]): PortModel[] {
    ports.forEach(value => {
      if (value.name.toLowerCase().includes('openvas')) {
        value.name = 'UTMStack Default';
      }
    });
    return ports.filter(value => !value.name.toLowerCase().includes('nmap'));
  }

  private onError(error) {
    // this.alertService.error(error.error, error.message, null);
  }

  private onSuccess(data, headers) {
    this.links = this.parseLinks.parse(headers.get('link'));
    this.totalItems = headers.get('X-Total-Count');
    this.ports = this.cleanPortRelatedToOpenvas(data);
    this.loading = false;
  }

}
