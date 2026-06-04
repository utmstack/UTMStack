import {Component, Input, OnInit} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {DataNatureTypeEnum} from '../../../../../shared/enums/nature-data.enum';
import {ElasticDataExportService} from '../../../../../shared/services/elasticsearch/elastic-data-export.service';
import {ElasticFilterType} from '../../../../../shared/types/filter/elastic-filter.type';
import {UtmFieldType} from '../../../../../shared/types/table/utm-field.type';
import {MAX_SEARCH_RESULTS} from "../../../../../shared/constants/global.constant";

@Component({
  selector: 'app-file-save-data',
  templateUrl: './file-save-data.component.html',
  styleUrls: ['./file-save-data.component.scss']
})
export class FileSaveDataComponent implements OnInit {
  @Input() filters: ElasticFilterType[];
  @Input() fields: UtmFieldType[];
  generateReport = false;
  limit: number;

  constructor(public activeModal: NgbActiveModal,
              private elasticDataExportService: ElasticDataExportService) {
  }

  ngOnInit() {
  }

  exportToCsv() {
    this.generateReport = true;
    const params = {
      columns: this.fields,
      dataOrigin: DataNatureTypeEnum.EVENT,
      filters: this.filters,
      top: MAX_SEARCH_RESULTS
    };
    this.elasticDataExportService.exportCsv(params, 'UTM FILE CLASSIFICATION').then(() => {
      this.generateReport = false;
    });
  }
}
