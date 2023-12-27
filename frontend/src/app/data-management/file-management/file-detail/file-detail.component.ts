import {Component, Input, OnInit} from '@angular/core';
import {UtmFieldType} from '../../../shared/types/table/utm-field.type';
import {getValueFromPropertyPath} from '../../../shared/util/get-value-object-from-property-path.util';
import {replaceBreakLine} from '../../../shared/util/string-util';
import {FILE_FIELDS} from '../shared/const/file-field.constant';
import {FileFieldEnum} from '../shared/enum/file-field.enum';

@Component({
  selector: 'app-file-detail',
  templateUrl: './file-detail.component.html',
  styleUrls: ['./file-detail.component.scss']
})
export class FileDetailComponent implements OnInit {
  @Input() file: any;
  fileFieldEnum = FileFieldEnum;
  fileFields: UtmFieldType[] = FILE_FIELDS;

  constructor() {
  }

  ngOnInit() {
  }

  getFieldByName(name): UtmFieldType {
    return FILE_FIELDS.find(value => value.field === name);
  }


  getEventMessage() {
    return replaceBreakLine(getValueFromPropertyPath(this.file, this.fileFieldEnum.FILE_MESSAGE_FIELD, null));
  }
}
