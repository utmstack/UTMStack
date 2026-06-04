import {Injectable} from '@angular/core';
import {ActivatedRouteSnapshot, CanActivate, Resolve, RouterStateSnapshot, Routes} from '@angular/router';
import {JhiResolvePagingParams} from 'ng-jhipster';
import {AccountService} from '../../core/auth/account.service';
import {User} from '../../core/user/user.model';
import {UserService} from '../../core/user/user.service';
import {UserMgmtDetailComponent} from './user-detail/user-management-detail.component';

import {UserMgmtComponent} from './user-list/user-management.component';
import {UserMgmtUpdateComponent} from './user-update/user-management-update.component';

@Injectable({providedIn: 'root'})
export class UserResolve implements CanActivate {
  constructor(private accountService: AccountService) {
  }

  canActivate() {
    return this.accountService.identity().then(account => this.accountService.hasAnyAuthority(['ROLE_ADMIN']));
  }
}

@Injectable({providedIn: 'root'})
export class UserMgmtResolve implements Resolve<any> {
  constructor(private service: UserService) {
  }

  resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot) {
    const id = route.params.login ? route.params.login : null;
    if (id) {
      return this.service.find(id);
    }
    return new User();
  }
}

export const userMgmtRoute: Routes = [
  {
    path: 'user',
    component: UserMgmtComponent,
    resolve: {
      pagingParams: JhiResolvePagingParams
    },
    data: {
      pageTitle: 'Users',
      defaultSort: 'id,asc'
    }
  },
  {
    path: 'user-management/:login/view',
    component: UserMgmtDetailComponent,
    resolve: {
      user: UserMgmtResolve
    },
    data: {
      pageTitle: 'Users'
    }
  },
  {
    path: 'user-management/new',
    component: UserMgmtUpdateComponent,
    resolve: {
      user: UserMgmtResolve
    }
  },
  {
    path: 'user-management/:login/edit',
    component: UserMgmtUpdateComponent,
    resolve: {
      user: UserMgmtResolve
    }
  }
];
