import {Component, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {NgbModalRef} from '@ng-bootstrap/ng-bootstrap';
import {ModalService} from '../../core/modal/modal.service';
import {LoginComponent} from '../../shared/components/auth/login/login.component';

import {ActivateService} from './activate.service';


@Component({
  selector: 'app-activate',
  templateUrl: './activate.component.html'
})
export class ActivateComponent implements OnInit {
  error: string;
  success: string;
  modalRef: NgbModalRef;

  constructor(private activateService: ActivateService,
              private loginModalService: ModalService,
              private route: ActivatedRoute) {
  }

  ngOnInit() {
    this.route.queryParams.subscribe(params => {
      this.activateService.get(params.key).subscribe(
        () => {
          this.error = null;
          this.success = 'OK';
        },
        () => {
          this.success = null;
          this.error = 'ERROR';
        }
      );
    });
  }

  login() {
    this.modalRef = this.loginModalService.open(LoginComponent);
  }
}
