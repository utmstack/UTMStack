import {Component, Input, OnInit} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {LOG_ANALYZER_TOTAL_ITEMS} from '../../../../../shared/constants/log-analyzer.constant';
import {DataNatureTypeEnum} from '../../../../../shared/enums/nature-data.enum';
import {ElasticDataExportService} from '../../../../../shared/services/elasticsearch/elastic-data-export.service';
import {ElasticFilterType} from '../../../../../shared/types/filter/elastic-filter.type';
import {UtmFieldType} from '../../../../../shared/types/table/utm-field.type';

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
      top: LOG_ANALYZER_TOTAL_ITEMS
    };
    this.elasticDataExportService.exportCsv(params, 'UTM FILE CLASSIFICATION').then(() => {
      this.generateReport = false;
    });
  }
}
