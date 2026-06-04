import {HttpResponse} from '@angular/common/http';
import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {delay} from 'rxjs/operators';
import {User} from '../../../../../core/user/user.model';
import {UserService} from '../../../../../core/user/user.service';
import {ITEMS_PER_PAGE} from '../../../../constants/pagination.constants';
import {SortEvent} from '../../../../directives/sortable/type/sort-event';
import {SortByType} from '../../../../types/sort-by.type';

@Component({
  selector: 'app-utm-user-select',
  templateUrl: './utm-user-select.component.html',
  styleUrls: ['./utm-user-select.component.css']
})
export class UtmUserSelectComponent implements OnInit {
  users: User[];
  error: any;
  success: any;
  totalItems: any;
  itemsPerPage = 5;
  page: any;
  search: string;
  fields: SortByType[] = [
    {
      fieldName: 'Login',
      field: 'login'
    }
  ];
  searchingUser = false;
  private sortBy: string;
  loading = true;
  usersSelected: User[] = [];
  @Output() userChange = new EventEmitter<User[]>();

  constructor(
    private userService: UserService,
  ) {
  }

  ngOnInit() {
    this.loadAll();
  }

  addToSelected(user: User) {
    const index = this.usersSelected.findIndex(value => value.id === user.id);
    if (index === -1) {
      this.usersSelected.push(user);
    } else {
      this.usersSelected.splice(index, 1);
    }
    this.userChange.emit(this.usersSelected);

  }

  isSelected(vis): boolean {
    return this.usersSelected.findIndex(value => value.id === vis.id) > -1;
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

  loadPage(page: number) {
    this.page = page;
    this.loadAll();
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

  private onSuccess(data, headers) {
    this.totalItems = headers.get('X-Total-Count');
    this.users = data;
    this.loading = false;
    this.searchingUser = false;
    this.users = this.users.filter(usr => {
      const usrHides = ['system', 'user', 'anonymoususer', 'fsclient'];
      return !usrHides.includes(usr.login, 0);
    });
  }

  private onError(error) {
  }
}
