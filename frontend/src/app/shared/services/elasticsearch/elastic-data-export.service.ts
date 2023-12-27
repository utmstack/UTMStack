import {Injectable} from '@angular/core';
import * as moment from 'moment';
import {UtmToastService} from '../../alert/utm-toast.service';
import {ElasticDataService} from './elastic-data.service';

@Injectable({
  providedIn: 'root'
})
export class ElasticDataExportService {

  constructor(private elasticDataService: ElasticDataService,
              private utmToastService: UtmToastService) {
  }

  exportCsv(params, filePrefix): Promise<boolean> {
    return new Promise<boolean>(resolve => {
      this.elasticDataService.exportToCsv(params).subscribe((dat) => {
        const data = new Blob([dat], {type: 'text/csv;charset=utf-8;'});
        if (data.size > 0) {
          // Browsers that support HTML5 download attribute
          const link = document.createElement('a');
          const url = URL.createObjectURL(data);
          link.setAttribute('href', url);
          const csvName = filePrefix + '-' + moment(new Date()).format('YYYY-MM-DD') + '.csv';
          link.setAttribute('download', csvName);
          link.style.visibility = 'hidden';
          document.body.appendChild(link);
          link.click();
          document.body.removeChild(link);
        } else {
          this.utmToastService.showWarning('NO DATA FOUND', 'No data found for this report');
        }
        resolve(true);
      });
    });
  }
}
