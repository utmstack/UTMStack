import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';
import {ActiveDirectoryTreeType} from '../types/active-directory-tree.type';

@Injectable({providedIn: 'root'})
export class TreeObjectBehavior {
  private user: BehaviorSubject<ActiveDirectoryTreeType> = new BehaviorSubject<ActiveDirectoryTreeType>(null);
  userSelected() {
    return this.user.asObservable();
  }

  changeUser(user: ActiveDirectoryTreeType) {
    this.user.next(user);
  }
}
