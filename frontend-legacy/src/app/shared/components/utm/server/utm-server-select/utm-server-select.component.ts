import {HttpResponse} from '@angular/common/http';
import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmServerService} from '../../../../../app-module/shared/services/utm-server.service';
import {UtmServerType} from '../../../../../app-module/shared/type/utm-server.type';

@Component({
  selector: 'app-utm-server-select',
  templateUrl: './utm-server-select.component.html',
  styleUrls: ['./utm-server-select.component.css']
})
export class UtmServerSelectComponent implements OnInit {
  @Input() server: UtmServerType;
  servers: UtmServerType[];
  @Input() selectFirst = false;
  @Input() showLabel = true;
  @Input() display: 'row' | 'column' = 'row';
  @Output() serverChange = new EventEmitter<UtmServerType>();

  constructor(public activeModal: NgbActiveModal,
              private modalService: NgbModal,
              private utmServerService: UtmServerService) {
  }

  ngOnInit() {
    this.getServers();
  }

  getServers(init?: string) {
    const req = {
      page: 0,
      size: 1000,
      sort: 'id,asc'
    };
    this.utmServerService.query(req).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers, init),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  onServerChange($event: any) {
    this.serverChange.emit($event);
  }

  private onError(error) {
    // this.alertService.error(error.error, error.message, null);
  }

  private onSuccess(data, headers, init) {
    this.servers = data;
    if (this.selectFirst) {
      this.server = this.servers[0];
      this.serverChange.emit(this.server);
    }
  }
}
