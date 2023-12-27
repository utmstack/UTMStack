import {HttpResponse} from '@angular/common/http';
import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {SortEvent} from '../../../../shared/directives/sortable/type/sort-event';
import {AssetSoftwareService} from '../../services/asset-software.service';
import {NetScanSoftwares} from '../../types/net-scan-softwares';
import {AssetSoftwareAddComponent} from '../asset-software-add/asset-software-add.component';

@Component({
  selector: 'app-asset-software-select',
  templateUrl: './asset-software-select.component.html',
  styleUrls: ['./asset-software-select.component.css']
})
export class AssetSoftwareSelectComponent implements OnInit {
  softwares: NetScanSoftwares[];
  loading = true;
  totalItems: any;
  page = 1;
  itemsPerPage = 10;
  searching = false;
  @Output() softwareSelected = new EventEmitter<NetScanSoftwares[]>();
  @Input() selected: NetScanSoftwares[] = [];
  private requestParams: any;
  private sortBy: SortEvent;

  constructor(private modalService: NgbModal,
              private assetSoftwareService: AssetSoftwareService,
              private utmToastService: UtmToastService) {
  }

  ngOnInit() {
    this.requestParams = {
      page: this.page - 1,
      size: this.itemsPerPage,
      sort: this.sortBy,
      'name.contains': null
    };
    this.getSoftwareList();
  }


  loadPage(page: any) {
    this.requestParams.page = page - 1;
    this.getSoftwareList();
  }

  getSoftwareList() {
    this.assetSoftwareService.query(this.requestParams).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  onSearchSoftware($event: string) {
    this.searching = true;
    this.requestParams['softName.contains'] = $event;
    this.getSoftwareList();
  }

  addSoftware() {
    const modal = this.modalService.open(AssetSoftwareAddComponent, {centered: true});
    modal.componentInstance.softwareCreated.subscribe(soft => {
      this.selected.push(soft);
      this.getSoftwareList();
    });
  }

  addToSelected(software: NetScanSoftwares) {
    const index = this.selected.findIndex(value => value.id === software.id);
    if (index === -1) {
      this.selected.push(software);
    } else {
      this.selected.splice(index, 1);
    }
    this.softwareSelected.emit(this.selected);
  }

  isSelected(software: NetScanSoftwares): boolean {
    return this.selected.findIndex(value => value.id === software.id) !== -1;
  }

  private onSuccess(data, headers) {
    this.totalItems = headers.get('X-Total-Count');
    this.softwares = data;
    this.loading = false;
    this.searching = false;
  }

  private onError(body: any) {
  }
}
