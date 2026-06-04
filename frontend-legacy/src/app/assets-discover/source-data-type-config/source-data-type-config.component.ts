import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {DataType} from '../../rule-management/models/rule.model';
import {DataTypeService} from '../../rule-management/services/data-type.service';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {ModalConfirmationComponent} from '../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';

@Component({
  selector: 'app-source-data-type-config',
  templateUrl: './source-data-type-config.component.html',
  styleUrls: ['./source-data-type-config.component.scss']
})
export class SourceDataTypeConfigComponent implements OnInit {
  @Output() refreshDataInput = new EventEmitter<boolean>();
  originals: DataType[] = [];
  dataInputs: DataType[] = [];
  saving = false;
  changed: DataType[] = [];
  loading = true;

  constructor(public activeModal: NgbActiveModal, private modalService: NgbModal,
              private dataTypeService: DataTypeService,
              private utmToastService: UtmToastService) {
  }

  ngOnInit() {
    this.getInputTypes();
  }

  getInputTypes() {
    this.dataTypeService.getAll({page: 0, size: 1000}).subscribe(response => {
      if (response.body) {
        this.originals = response.body;
        this.loadElements().then(data => {
          this.dataInputs = data;
          this.loading = false;
        });
      }
    });
  }

  loadElements(): Promise<DataType[]> {
    return new Promise<DataType[]>(resolve => {
      const inputs: DataType[] = [];
      for (const data of this.originals) {
        inputs.push(data);
      }
      resolve(inputs);
    });
  }

  search($event: string) {
    if (!$event) {
      this.dataInputs = this.originals;
    } else {
      this.dataInputs = this.originals.filter(value =>
        value.dataTypeName.toLowerCase().includes($event.toLowerCase())
      ).slice();
    }
  }

  update() {
    const modalSource = this.modalService.open(ModalConfirmationComponent, {centered: true});
    const changedStr = this.changed.map(value => value.dataTypeName);
    let result = '';
    if (changedStr.length > 1) {
      const lastItem = changedStr.pop();
      result = changedStr.join(', ') + ', and ' + lastItem;
    } else {
      result = changedStr.join(', ');
    }
    modalSource.componentInstance.header = 'Disable source input view';
    modalSource.componentInstance.message = 'Are you sure that you want to hidde the sources ' + result + '?';
    modalSource.componentInstance.confirmBtnText = 'Accept';
    modalSource.componentInstance.confirmBtnIcon = 'icon-cog3';
    modalSource.componentInstance.confirmBtnType = 'default';
    modalSource.result.then(() => {
      this.save();
    });
  }

  save() {
    this.saving = true;
    this.dataTypeService.updateInclude(this.changed).subscribe(() => {
      this.saving = false;
      this.refreshDataInput.emit(true);
      this.utmToastService.showSuccessBottom('Configuration saved successfully, your sources will be updated');
      for (const sourceDataType of this.changed) {
        const indexOriginal = this.originals.findIndex(value => value.dataType === sourceDataType.dataType);
        const indexChanged = this.changed.findIndex(value => value.dataType === sourceDataType.dataType);
        this.originals[indexOriginal] = this.changed[indexChanged];
        this.changed = [];
      }

    }, () => {
      this.utmToastService.showError('Error', 'Error trying update source configuration');
      this.saving = false;
    });

  }

  changeStatus(sourceDataType: DataType) {
    sourceDataType.included = !sourceDataType.included;
    const indexChanged = this.changed.findIndex(value => value.dataType === sourceDataType.dataType);

    if (indexChanged !== -1) {
      this.changed[indexChanged].included = sourceDataType.included;
    } else {
      this.changed.push(sourceDataType);
    }
  }

}
