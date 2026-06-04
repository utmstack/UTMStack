import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../../shared/alert/utm-toast.service';
import {ITEMS_PER_PAGE} from '../../../../../shared/constants/pagination.constants';
import {SortByType} from '../../../../../shared/types/sort-by.type';
import {InputClassResolve} from '../../../../../shared/util/input-class-resolve';
import {PortModel} from '../../../../shared/model/port.model';
import {PortService} from '../../shared/services/port.service';
import {PortRangeCreateComponent} from '../port-range-create/port-range-create.component';
import {PortRangeDeleteComponent} from '../port-range-delete/port-range-delete.component';

@Component({
  selector: 'app-port-range-list',
  templateUrl: './port-range-list.component.html',
  styleUrls: ['./port-range-list.component.scss']
})
export class PortRangeListComponent implements OnInit {
  @Input() port: PortModel;
  @Output() portEdited = new EventEmitter<string>();
  fields: SortByType[] = [
    {
      fieldName: 'Type',
      field: 'type'
    },
    {
      fieldName: 'Start',
      field: 'start'
    },
    {
      fieldName: 'End',
      field: 'end'
    }];
  loading = false;
  totalItems: number;
  page: any;
  itemsPerPage: any;
  formPortList: FormGroup;

  constructor(private portListService: PortService,
              private fb: FormBuilder,
              public inputClass: InputClassResolve,
              public activeModal: NgbActiveModal,
              private modalService: NgbModal,
              private utmToastService: UtmToastService) {
  }

  ngOnInit() {
    this.itemsPerPage = ITEMS_PER_PAGE;
    this.totalItems = this.port.portRanges.length;
    this.page = 0;
    // this.getPort();
    this.initFormPort();
    this.setForm();
  }


  initFormPort() {
    this.formPortList = this.fb.group({
      name: ['', [Validators.required]],
      comment: ['', Validators.required],
      portListId: [this.port.uuid]
    });
  }

  setForm() {
    this.formPortList.get('name').setValue(this.port.name);
    this.formPortList.get('comment').setValue(this.port.comment);
  }

  getPort() {
    const req = {
      'uuid.equals': this.port.uuid,
      details: true
    };
    this.portListService.query(req).subscribe(portDetail => {
      this.port = portDetail.body.portList[0];
    });
  }

  onSortBy($event: string) {
  }

  loadPage(page: any) {
  }

  // paginator(items, page, per_page) {
  //
  //   var page = page || 1,
  //     per_page = per_page || 10,
  //     offset = (page - 1) * per_page,
  //
  //     paginatedItems = items.slice(offset).slice(0, per_page),
  //     total_pages = Math.ceil(items.length / per_page);
  //   return {
  //     page: page,
  //     per_page: per_page,
  //     pre_page: page - 1 ? page - 1 : null,
  //     next_page: (total_pages > page) ? page + 1 : null,
  //     total: items.length,
  //     total_pages: total_pages,
  //     data: paginatedItems
  //   };

  savePortList() {
    this.portListService.update(this.formPortList.value).subscribe(portCreated => {
      this.portEdited.emit('success');
      this.activeModal.dismiss();
      this.utmToastService.showSuccessBottom('Port edited successfully');
    }, error1 => {
      this.utmToastService.showError('Error editing port',
        error1.error.statusText);
    });
  }

  addPortRange() {
    const modal = this.modalService.open(PortRangeCreateComponent, {centered: true});
    modal.componentInstance.port = this.port;
    modal.componentInstance.portRangeAdded.subscribe(() => {
      this.getPort();
    });
  }

  deletePortRange(portRange: PortModel) {
    const modal = this.modalService.open(PortRangeDeleteComponent, {centered: true});
    modal.componentInstance.portRange = portRange;
    modal.componentInstance.portRangeDeleted.subscribe(() => {
      this.getPort();
    });
  }
}

