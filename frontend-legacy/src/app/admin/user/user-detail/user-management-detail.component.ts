import {Component, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';

import {User} from '../../../core/user/user.model';

@Component({
  selector: 'app-user-mgmt-detail',
  templateUrl: './user-management-detail.component.html'
})
export class UserMgmtDetailComponent implements OnInit {
  user: User;

  constructor(private route: ActivatedRoute) {
  }

  ngOnInit() {
    this.route.data.subscribe(({user}) => {
      this.user = user.body ? user.body : user;
    });
  }
}
