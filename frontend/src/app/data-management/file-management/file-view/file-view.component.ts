import {HttpResponse} from '@angular/common/http';
import {Component, OnInit} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {ResizeEvent} from 'angular-resizable-element';
import {ElasticFilterDefaultTime} from '../../../shared/components/utm/filters/elastic-filter-time/elastic-filter-time.component';
import {FILE_ROUTE} from '../../../shared/constants/app-routes.constant';
import {LOG_INDEX_WINLOGBEAT} from '../../../shared/constants/main-index-pattern.constant';
import {ITEMS_PER_PAGE} from '../../../shared/constants/pagination.constants';
import {SortEvent} from '../../../shared/directives/sortable/type/sort-event';
import {ElasticOperatorsEnum} from '../../../shared/enums/elastic-operators.enum';
import {ElasticDataService} from '../../../shared/services/elasticsearch/elastic-data.service';
import {ElasticFilterType} from '../../../shared/types/filter/elastic-filter.type';
import {UtmFieldType} from '../../../shared/types/table/utm-field.type';
import {TimeFilterType} from '../../../shared/types/time-filter.type';
import {sanitizeFilters} from '../../../shared/util/elastic-filter.util';
import {FileFiltersBehavior} from '../shared/behavior/file-filters.behavior';
import {FileSaveDataComponent} from '../shared/components/file-save-data/file-save-data.component';
import {
  ALL_FILE_EVENT_ID_NUMBER,
  CREATED_FILE_EVENT_ID_NUMBER,
  DELETED_FILE_EVENT_ID_NUMBER,
  FILE_COMMON_FIELDS,
  FILE_FIELDS,
  FILE_OBJECT_TYPE_VALUE,
  FILE_PERMISSION_FIELDS,
  FILE_SHARED_FIELDS,
  PERMISSION_FILE_EVENT_ID_NUMBER,
  SHARE_FILE_EVENT_ID_NUMBER
} from '../shared/const/file-field.constant';
import {AccessMaskEnum} from '../shared/enum/access-mask.enum';
import {FileFieldEnum} from '../shared/enum/file-field.enum';
import {FileQueryParamEnum} from '../shared/enum/file-query-param.enum';

@Component({
  selector: 'app-file-view',
  templateUrl: './file-view.component.html',
  styleUrls: ['./file-view.component.scss']
})
export class FileViewComponent implements OnInit {
  files: any[] = [];
  fields = FILE_COMMON_FIELDS;
  defaultTime: ElasticFilterDefaultTime = new ElasticFilterDefaultTime('now-7d', 'now');
  pageWidth = window.innerWidth;
  filterWidth: number;
  tableWidth: number;
  sortEvent: any;
  totalItems: any;
  page = 1;
  loading = true;
  itemsPerPage = ITEMS_PER_PAGE;
  viewFileDetail: any;
  filters: ElasticFilterType[];
  sortBy = FileFieldEnum.FILE_TIMESTAMP_FIELD + ',desc';
  fileTitle = 'File classification';

  constructor(private elasticDataService: ElasticDataService,
              private router: Router,
              private modalService: NgbModal,
              private activeRoute: ActivatedRoute,
              private fileFiltersBehavior: FileFiltersBehavior) {

  }

  ngOnInit() {
    this.setInitialWidth();
    this.activeRoute.queryParams.subscribe((params: any) => {
      this.setFiltersByFileType(params.filterBy).then(filters => {
        this.filters = filters;
        this.fileFiltersBehavior.$fileFilters.next(this.filters);
        this.getFiles();
      });
    });
  }

  setFiltersByFileType(fileFilter: FileQueryParamEnum): Promise<ElasticFilterType[]> {
    return new Promise<ElasticFilterType[]>(resolve => {
      const filters: ElasticFilterType[] = [
        {field: FileFieldEnum.FILE_OBJECT_TYPE_FIELD, operator: ElasticOperatorsEnum.IS_ONE_OF, value: FILE_OBJECT_TYPE_VALUE},
        {field: FileFieldEnum.FILE_TIMESTAMP_FIELD, operator: ElasticOperatorsEnum.IS_BETWEEN, value: ['now-7d', 'now']}
      ];
      switch (fileFilter) {
        case FileQueryParamEnum.ALL_EVENTS:
          this.fileTitle = 'Event explorer';
          this.fields = FILE_FIELDS;
          filters.push({
            field: FileFieldEnum.FILE_EVENT_ID_FIELD,
            operator: ElasticOperatorsEnum.IS_ONE_OF,
            value: ALL_FILE_EVENT_ID_NUMBER
          });
          break;
        case FileQueryParamEnum.FILES_CREATED:
          this.fileTitle = 'Files created';
          filters.push({
            field: FileFieldEnum.FILE_EVENT_ID_FIELD,
            operator: ElasticOperatorsEnum.IS,
            value: CREATED_FILE_EVENT_ID_NUMBER
          });
          filters.push({
            field: FileFieldEnum.FILE_ACCESS_MASK_FIELD,
            operator: ElasticOperatorsEnum.IS_ONE_OF,
            value: [AccessMaskEnum.WRITE_DATA, AccessMaskEnum.APPEND_DATA]
          });
          break;
        case FileQueryParamEnum.FILES_DELETED:
          this.fileTitle = 'Files deleted';
          filters.push({
            field: FileFieldEnum.FILE_EVENT_ID_FIELD,
            operator: ElasticOperatorsEnum.IS_ONE_OF,
            value: DELETED_FILE_EVENT_ID_NUMBER
          });
          filters.push({
            field: FileFieldEnum.FILE_ACCESS_MASK_FIELD,
            operator: ElasticOperatorsEnum.IS_ONE_OF,
            value: [AccessMaskEnum.DELETE, AccessMaskEnum.DELETE_CHILD]
          });
          break;
        case FileQueryParamEnum.FILES_MODIFIED:
          this.fileTitle = 'Files modified';
          break;
        case FileQueryParamEnum.FILES_MOVED:
          this.fileTitle = 'Files moved';
          break;
        case FileQueryParamEnum.FILES_RENAMED:
          this.fileTitle = 'Files Renamed';
          break;
        case FileQueryParamEnum.PERMISSION_CHANGE:
          this.fileTitle = 'Files permission changes';
          this.fields = FILE_PERMISSION_FIELDS;
          filters.push({
            field: FileFieldEnum.FILE_EVENT_ID_FIELD,
            operator: ElasticOperatorsEnum.IS,
            value: PERMISSION_FILE_EVENT_ID_NUMBER
          });
          break;
        case FileQueryParamEnum.FILES_SHARED:
          this.fields = FILE_SHARED_FIELDS;
          this.fileTitle = 'Shared files';
          filters.push({
            field: FileFieldEnum.FILE_EVENT_ID_FIELD,
            operator: ElasticOperatorsEnum.IS_ONE_OF,
            value: SHARE_FILE_EVENT_ID_NUMBER
          });
          break;
      }
      resolve(filters);

    });
  }

  getFiles(calledFrom?: string) {
    this.elasticDataService.search(this.page, this.itemsPerPage,
      100000000, LOG_INDEX_WINLOGBEAT,
      sanitizeFilters(this.filters), this.sortBy).subscribe(
      (res: HttpResponse<any>) => {
        this.totalItems = Number(res.headers.get('X-Total-Count'));
        this.files = res.body;
        this.loading = false;
      },
      (res: HttpResponse<any>) => {
      }
    );
  }

  onTimeFilterChange($event: TimeFilterType) {
    const timeFilterIndex = this.filters.findIndex(value => value.field === FileFieldEnum.FILE_TIMESTAMP_FIELD);
    if (timeFilterIndex === -1) {
      this.filters.push({
        field: FileFieldEnum.FILE_TIMESTAMP_FIELD,
        value: [$event.timeFrom, $event.timeTo],
        operator: ElasticOperatorsEnum.IS_BETWEEN
      });
    } else {
      this.filters[timeFilterIndex].value = [$event.timeFrom, $event.timeTo];
    }
    this.fileFiltersBehavior.$fileFilters.next(this.filters);
    this.getFiles('on time filter change');
  }


  setInitialWidth() {
    if (this.pageWidth > 1980) {
      this.filterWidth = 350;
      this.tableWidth = this.pageWidth - this.filterWidth - 51;
    } else {
      this.filterWidth = 300;
      this.tableWidth = this.pageWidth - this.filterWidth - 51;
    }
    if (this.pageWidth > 2500) {
      this.filterWidth = 350;
      this.tableWidth = this.pageWidth - this.filterWidth - 51;
    }
    if (this.pageWidth > 4000) {
      this.filterWidth = 400;
      this.tableWidth = this.pageWidth - this.filterWidth - 51;
    }
  }

  saveReport() {
    const modalSaveReport = this.modalService.open(FileSaveDataComponent, {centered: true});
    modalSaveReport.componentInstance.filters = this.filters;
    modalSaveReport.componentInstance.fields = this.fields.filter(value => value.visible === true);
  }

  onSortBy($event: SortEvent) {
    this.sortBy = $event.column + ',' + $event.direction;
    this.getFiles('on sort by');
  }

  addToSelected(file: any) {

  }

  isSelected(file: any) {
    return false;
  }

  loadPage($event: number) {
    this.page = $event;
    this.getFiles();
  }

  onItemsPerPageChange($event: number) {
    this.itemsPerPage = $event;
    this.getFiles();
  }

  onResize($event: ResizeEvent) {
    if ($event.rectangle.width >= 250) {
      this.tableWidth = (this.pageWidth - $event.rectangle.width - 51);
      this.filterWidth = $event.rectangle.width;
    }
  }

  onDataRefresh($event: boolean) {

  }

  onFilterChange($event: ElasticFilterType) {
    this.processFileFilters($event).then(filters => {
      this.filters = filters;
      this.fileFiltersBehavior.$fileFilters.next(this.filters);
      this.getFiles('on generic filter change');
    });
  }

  processFileFilters(filter: ElasticFilterType): Promise<ElasticFilterType[]> {
    return new Promise<ElasticFilterType[]>(resolve => {
      const indexFilters = this.filters.findIndex(value => filter.field.includes(value.field));
      if (indexFilters === -1) {
        this.filters.push(filter);
      } else {
        this.filters[indexFilters] = filter;
      }
      resolve(this.filters);
    });
  }

  onFilterReset($event: boolean) {
    this.router.navigate([FILE_ROUTE]);
  }

  onFilterAppliedChange($event: { filter: ElasticFilterType, valueDelete: string }) {
    this.processFileFilters($event.filter).then(filters => {
      this.fileFiltersBehavior.$fileDeleteFilterValue.next({field: $event.filter.field, value: $event.valueDelete});
      this.filters = filters;
      this.getFiles('on filter applied change');
    });
  }

  viewDetail(file: any, td: UtmFieldType) {
    if (td.field !== FileFieldEnum.FILE_NEW_SDDL_FIELD && td.field !== FileFieldEnum.FILE_OLD_SDDL_FIELD) {
      this.viewFileDetail = file;
    }
  }
}
