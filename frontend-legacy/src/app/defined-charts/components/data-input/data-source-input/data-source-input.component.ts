import {HttpResponse} from '@angular/common/http';
import {Component, OnDestroy, OnInit, QueryList, ViewChildren} from '@angular/core';
import {SortableDirective} from '../../../../shared/directives/sortable/sortable.directive';
import {SortEvent} from '../../../../shared/directives/sortable/type/sort-event';
import {DataSourceInputService} from './services/data-source-input.service';
import {UtmDataInputStatus} from './type/data-source-input.type';

@Component({
  selector: 'app-data-source-input',
  templateUrl: './data-source-input.component.html',
  styleUrls: ['./data-source-input.component.css']
})
export class DataSourceInputComponent implements OnInit, OnDestroy {
  data: DataSourceInput[] = [];
  loadingOption = true;
  @ViewChildren(SortableDirective) headers: QueryList<SortableDirective>;
  interval: any;
  private direction: 'asc' | 'desc' | '';
  private responseRows: DataSourceInput[] = [];
  deleting: string[] = [];
  totalItems: number;
  page = 1;
  itemsPerPage = 10;
  pageStart = 0;
  pageEnd = 10;

  constructor(private dataSourceInputService: DataSourceInputService) {
  }

  ngOnInit() {
    this.getSourceStatus();
    this.interval = setInterval(() => this.getSourceStatus(), 15000);
  }

  ngOnDestroy(): void {
    clearInterval(this.interval);
  }

  getSourceStatus() {
    this.dataSourceInputService.query({page: 0, size: 100000}).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  onSort({column, direction}: SortEvent) {
    this.direction = direction;
    this.headers.forEach(header => {
      if (header.sortable !== column) {
        header.direction = '';
      }
    });
    if (direction === '') {
      this.data = this.responseRows;
    } else {
      this.data = this.data.sort((a, b) => {
        if (direction === 'asc') {
          return a[column] > b[column] ? 1 : -1;
        } else {
          return a[column] < b[column] ? 1 : -1;
        }
      });
    }
  }

  convertToDate(row: DataSourceInput) {
    return new Date(Number(row.lastInput) * 1000);
  }

  private onError(error) {
    // this.alertService.error(error.error, error.message, null);
  }

  groupByHost(data: UtmDataInputStatus[]): Promise<DataSourceInput[]> {
    return new Promise<DataSourceInput[]>(resolve => {
      let dataSource: DataSourceInput[] = [];
      data.forEach(value => {
        const indexSource = dataSource.findIndex(valueDatasource => valueDatasource.host === value.source);
        if (indexSource === -1) {
          dataSource.push({
            host: value.source,
            dataType: [{isDown: value.down, type: value.dataType, id: value.id}],
            hostIsDown: value.down,
            lastInput: value.timestamp,
          });
        } else {
          dataSource[indexSource].dataType.push({isDown: value.down, type: value.dataType, id: value.id});
          dataSource[indexSource].hostIsDown = dataSource[indexSource].dataType
            .findIndex(val => val.isDown) !== -1;
        }
      });
      dataSource = dataSource.sort((a, b) => a.host < b.host ? 1 : -1);
      resolve(dataSource);
    });
  }

  getDisconnected(dataType: { isDown: boolean; type: string }[]): number {
    return dataType.filter(value => value.isDown).length;
  }

  private onSuccess(data, headers) {
    this.totalItems = headers.get('X-Total-Count');
    this.groupByHost(data).then(group => {
      this.data = group;
      this.responseRows = group;
    });
    this.loadingOption = false;
  }

  delete(row: DataSourceInput, dat: { isDown: boolean; type: string; id?: string }) {
    this.deleting.push(dat.id);
    this.dataSourceInputService.delete(dat.id).subscribe(() => {
      const index = this.data.indexOf(row);
      if (index !== -1) {
        const idIndex = this.data[index].dataType.indexOf(dat);
        if (idIndex !== -1) {
          this.data[index].dataType.splice(idIndex, 1);
        }
      }
      const indexDelete = this.deleting.indexOf(dat.id);
      if (indexDelete !== -1) {
        this.deleting.splice(indexDelete, 1);
      }
    });
  }

  onSearch($event: string) {
    if ($event) {
      this.data = this.data.filter(value => value.host.toLowerCase().includes($event.toLowerCase()));
    } else {
      this.data = this.responseRows;
    }
  }

  loadPage($event: number) {
    this.pageEnd = this.page * this.itemsPerPage;
    this.pageStart = this.pageEnd - this.itemsPerPage;
  }
}


export class DataSourceInput {
  hostIsDown: boolean;
  host: string;
  dataType: { isDown: boolean, type: string, id?: string }[];
  lastInput: Date;
}
