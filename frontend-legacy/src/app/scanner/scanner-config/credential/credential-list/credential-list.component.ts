import {HttpResponse} from '@angular/common/http';
import {Component, OnInit} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {ITEMS_PER_PAGE} from '../../../../shared/constants/pagination.constants';
import {SortEvent} from '../../../../shared/directives/sortable/type/sort-event';
import {SortByType} from '../../../../shared/types/sort-by.type';
import {UsedByComponent} from '../../../shared/components/used-by/used-by.component';
import {CredentialModel} from '../../../shared/model/credential.model';
import {CredentialCreateComponent} from '../credential-create/credential-create.component';
import {CredentialDeleteComponent} from '../credential-delete/credential-delete.component';
import {CredentialService} from '../shared/services/credential.service';

@Component({
  selector: 'app-credential-list',
  templateUrl: './credential-list.component.html',
  styleUrls: ['./credential-list.component.scss']
})
export class CredentialListComponent implements OnInit {
  credentials: CredentialModel[] = [];
  loading = false;
  totalItems: any;
  page = 1;
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
      fieldName: 'Name',
      field: 'name'
    },
    {
      fieldName: 'Type',
      field: 'type'
    },
    {
      fieldName: 'Allow insecure use',
      field: 'allow_insecure'
    }
  ];
  credentialName: string;
  insecure = -1;
  private sortBy: string;

  constructor(private modalService: NgbModal,
              private activatedRoute: ActivatedRoute,
              private router: Router,
              private credentialService: CredentialService) {
  }

  ngOnInit() {
    this.request = {
      page: this.page - 1,
      size: this.itemsPerPage,
      sort: this.sortBy,
      targets: true
    };
    this.loadCredential();
  }

  loadCredential() {
    this.getCredentials();
  }

  newCredential() {
    const modal = this.modalService.open(CredentialCreateComponent, {centered: true});
    modal.componentInstance.credentialCreated.subscribe(created => {
      this.getCredentials();
    });
  }

  editCredential(credential: any) {
    const modal = this.modalService.open(CredentialCreateComponent, {centered: true});
    modal.componentInstance.credential = credential;
    modal.componentInstance.credentialCreated.subscribe(created => {
      this.getCredentials();
    });
  }

  deleteCredential(credential: CredentialModel) {
    if (credential.inUse === '0') {
      const modal = this.modalService.open(CredentialDeleteComponent, {centered: true});
      modal.componentInstance.credential = credential;
      modal.componentInstance.credentialDeleted.subscribe(() => {
        this.getCredentials();
      });
    }
  }

  onSortBy($event: SortEvent) {
    this.sortBy = $event.column + ',' + $event.direction;
    this.request.sort = $event.column + ',' + $event.direction;
    this.getCredentials();
  }


  getCredentials() {
    this.loading = true;
    this.credentials = [];
    this.credentialService.query(this.request).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  loadPage(page: number) {
    if (page !== this.previousPage) {
      this.previousPage = page;
      this.transition();
    }
  }

  transition() {
    this.router.navigate(['/scanner/config/credential'], {
      queryParams: {
        page: this.page,
        size: this.itemsPerPage,
        sort: this.sortBy
      }
    });
    this.getCredentials();
  }

  showUse(credential: CredentialModel) {
    const modal = this.modalService.open(UsedByComponent, {centered: true});
    modal.componentInstance.using = credential.targets;
    modal.componentInstance.dependency = 'Targets';
    modal.componentInstance.type = 'credential';
    modal.componentInstance.name = credential.name;
  }

  filterByName() {
    if (this.credentialName !== '') {
      this.request['name.contains'] = this.credentialName;
    } else {
      this.request['name.contains'] = undefined;
    }
    setTimeout(() => this.loadCredential(), 1000);
  }

  allowInsecure() {
    if (this.insecure !== -1) {
      this.request['allowInsecure.equals'] = this.insecure;
      this.loadCredential();
    } else {
      this.request['allowInsecure.equals'] = undefined;
      this.loadCredential();
    }
  }

  private onSuccess(data, headers) {
    this.totalItems = headers.get('X-Total-Count');
    this.credentials = data;
    this.loading = false;
  }

  private onError(error) {
    // this.alertService.error(error.error, error.message, null);
  }
}
