import {Component, Input, OnInit} from '@angular/core';
import {getValueFromPropertyPath} from '../../../../../shared/util/get-value-object-from-property-path.util';
import {SDDLParser} from '../../../../../shared/util/sddl-parser.util';
import {FileFieldEnum} from '../../enum/file-field.enum';

@Component({
  selector: 'app-file-sddl-view',
  templateUrl: './file-sddl-view.component.html',
  styleUrls: ['./file-sddl-view.component.scss']
})
export class FileSddlViewComponent implements OnInit {
  @Input() file: any;
  fileFieldEnum = FileFieldEnum;
  sddl: any;

  constructor() {
  }

  ngOnInit() {
    const fileSddl = getValueFromPropertyPath(this.file, this.fileFieldEnum.FILE_NEW_SDDL_FIELD, null);
    if (fileSddl) {
      this.sddl = new SDDLParser().parse(fileSddl);
    }
  }

}
