import {Component, OnInit} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {UtmClientService} from '../../shared/services/license/utm-client.service';
import {UtmClientType} from '../../shared/types/license/utm-client.type';
import {UtmLicenseActivateComponent} from '../shared/components/utm-license-activate/utm-license-activate.component';

@Component({
  selector: 'app-utm-license',
  templateUrl: './utm-license.component.html',
  styleUrls: ['./utm-license.component.scss']
})
export class UtmLicenseComponent implements OnInit {
  client: UtmClientType;

  constructor(private utmClientService: UtmClientService,
              private toastService: UtmToastService,
              private modalService: NgbModal) {
  }

  ngOnInit() {
    this.getClient();
  }

  getClient() {
    this.utmClientService.query({size: 100, page: 0}).subscribe(response => {
      this.client = response.body[0];
    });
  }

  showActivateModal() {
    const modal = this.modalService.open(UtmLicenseActivateComponent, {centered: true});
    modal.componentInstance.licenseChecked.subscribe(check => {
      this.getClient();
    });
  }
}
