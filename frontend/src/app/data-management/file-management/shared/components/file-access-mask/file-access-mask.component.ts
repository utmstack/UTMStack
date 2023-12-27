import {Component, Input, OnInit} from '@angular/core';
import {getValueFromPropertyPath} from '../../../../../shared/util/get-value-object-from-property-path.util';
import {replaceBreakLine} from '../../../../../shared/util/string-util';
import {ACCESS_MASK_CODES} from '../../const/file-acces-mask.constant';
import {FileFieldEnum} from '../../enum/file-field.enum';
import {FileAccessMaskCodeType} from '../../types/file-access-mask-code.type';

@Component({
  selector: 'app-file-access-mask',
  templateUrl: './file-access-mask.component.html',
  styleUrls: ['./file-access-mask.component.scss']
})
export class FileAccessMaskComponent implements OnInit {
  @Input() file: any;
  fileFieldEnum = FileFieldEnum;
  accessMasksCodes = ACCESS_MASK_CODES;
  access: FileAccessMaskCodeType;

  constructor() {
  }

  ngOnInit() {
    const accessMask = getValueFromPropertyPath(this.file, this.fileFieldEnum.FILE_ACCESS_MASK_FIELD, null);
    if (accessMask) {
      this.access = this.accessMasksCodes.find(value => value.hex === accessMask);
    }
  }

  convertDescToHtml(): string {
    return replaceBreakLine(this.access.description);
  }
}
