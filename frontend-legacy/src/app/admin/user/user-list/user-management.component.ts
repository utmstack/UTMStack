import {HttpResponse} from '@angular/common/http';
import {Component, OnDestroy, OnInit} from '@angular/core';

import {ActivatedRoute, Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {JhiAlertService, JhiEventManager, JhiParseLinks} from 'ng-jhipster';
import {delay} from 'rxjs/operators';
import {AccountService} from '../../../core/auth/account.service';
import {User} from '../../../core/user/user.model';
import {UserService} from '../../../core/user/user.service';
import {ContactUsComponent} from '../../../shared/components/contact-us/contact-us.component';
import {DEMO_URL} from '../../../shared/constants/global.constant';
import {ITEMS_PER_PAGE} from '../../../shared/constants/pagination.constants';
import {SortEvent} from '../../../shared/directives/sortable/type/sort-event';
import {SortByType} from '../../../shared/types/sort-by.type';
import {UserMgmtDeleteDialogComponent} from '../user-delete/user-management-delete-dialog.component';
import {UserMgmtUpdateComponent} from '../user-update/user-management-update.component';

@Component({
  selector: 'app-user-mgmt',
  templateUrl: './user-management.component.html',
  styleUrls: ['./user-management.component.scss']
})
export class UserMgmtComponent implements OnInit, OnDestroy {
  currentAccount: any;
  users: User[];
  error: any;
  success: any;
  routeData: any;
  links: any;
  totalItems: any;
  itemsPerPage: any;
  page: any;
  predicate: any;
  previousPage: any;
  reverse: any;
  search: string;
  fields: SortByType[] = [
    {
      fieldName: 'Default',
      field: 'id'
    },
    {
      fieldName: 'Login',
      field: 'login'
    },
    {
      fieldName: 'Email',
      field: 'email'
    },
    {
      fieldName: 'Activated',
      field: 'activated'
    },
    {
      fieldName: 'Created Date',
      field: 'createdDate'
    },
    {
      fieldName: 'Last Modified By',
      field: 'lastModifiedBy'
    },
    {
      fieldName: 'Last Modified Date',
      field: 'lastModifiedDate'
    }

  ];
  searchingUser = false;
  private sortBy: string;

  constructor(
    private userService: UserService,
    private alertService: JhiAlertService,
    private accountService: AccountService,
    private parseLinks: JhiParseLinks,
    private activatedRoute: ActivatedRoute,
    private router: Router,
    private eventManager: JhiEventManager,
    private modalService: NgbModal
  ) {
    this.itemsPerPage = ITEMS_PER_PAGE;
    this.routeData = this.activatedRoute.data.subscribe(data => {
      this.page = data.pagingParams.page;
      this.previousPage = data.pagingParams.page;
      this.reverse = data.pagingParams.ascending;
      this.predicate = data.pagingParams.predicate;
      this.sortBy = this.predicate + ',' + (this.reverse ? 'asc' : 'desc');
    });
  }

  ngOnInit() {
    this.accountService.identity().then(account => {
      this.currentAccount = account;
      this.loadAll();
      this.registerChangeInUsers();
    });
  }

  ngOnDestroy() {
    this.routeData.unsubscribe();
  }

  registerChangeInUsers() {
    this.eventManager.subscribe('userListModification', response => this.loadAll());
  }

  setActive(user, isActivated) {
    user.activated = isActivated;

    this.userService.update(user).subscribe(response => {
      if (response.status === 200) {
        this.error = null;
        this.success = 'OK';
        this.loadAll();
      } else {
        this.success = null;
        this.error = 'ERROR';
      }
    });
  }

  loadAll() {
    this.userService
      .query({
        page: this.page - 1,
        size: this.itemsPerPage,
        sort: this.sortBy
      })
      .subscribe(
        (res: HttpResponse<User[]>) => this.onSuccess(res.body, res.headers),
        (res: HttpResponse<any>) => this.onError(res.body)
      );
  }

  trackIdentity(index, item: User) {
    return item.id;
  }

  sort() {
    const result = [this.predicate + ',' + (this.reverse ? 'asc' : 'desc')];
    if (this.predicate !== 'id') {
      result.push('id');
    }
    return result;
  }

  loadPage(page: number) {
    if (page !== this.previousPage) {
      this.previousPage = page;
      this.transition();
    }
  }

  transition() {
    this.router.navigate(['/management/user'], {
      queryParams: {
        page: this.page,
        sort: this.predicate + ',' + (this.reverse ? 'asc' : 'desc')
      }
    });
    this.loadAll();
  }

  deleteUser(user: User) {
    if (!window.location.href.includes(DEMO_URL)) {
      const modalRef = this.modalService.open(UserMgmtDeleteDialogComponent, {
        size: 'sm',
        backdrop: 'static',
        centered: true
      });
      modalRef.componentInstance.user = user;
      modalRef.result.then(
        result => {
          // Left blank intentionally, nothing to do here
        },
        reason => {
          // Left blank intentionally, nothing to do here
        }
      );
    } else {
      this.modalService.open(ContactUsComponent, {centered: true});
    }
  }

  searchUser($event) {
    this.search = $event;
    if (this.search !== '' || this.search !== undefined) {
      const req = {login: this.search};
      this.searchingUser = true;
      this.userService.query(req).pipe(delay(2000)).subscribe(users => {
        this.users = users.body;
        this.searchingUser = false;
      });
    } else {
      this.loadAll();
    }
  }

  onSort($event: SortEvent) {
    this.sortBy = $event.column + ',' + $event.direction;
    this.loadAll();
  }

  createUser() {
    const modal = this.modalService.open(UserMgmtUpdateComponent, {centered: true});
    modal.componentInstance.userChange.subscribe(() => {
      this.loadAll();
    });
  }

  editUser(user: User) {
    const modal = this.modalService.open(UserMgmtUpdateComponent, {centered: true});
    modal.componentInstance.user = user;
    modal.componentInstance.userChange.subscribe(() => {
      this.loadAll();
    });
  }

  private onSuccess(data, headers) {
    this.links = this.parseLinks.parse(headers.get('link'));
    this.totalItems = headers.get('X-Total-Count');
    this.users = data;
    this.searchingUser = false;
    this.users = this.users.filter(usr => {
      const usrHides = ['system', 'user'];
      return !usrHides.includes(usr.login, 0);
    });
  }

  private onError(error) {
    this.alertService.error(error.error, error.message, null);
  }
}
