import {Component, Input, OnInit} from '@angular/core';
import {
  UtmTableDetailComponent
} from '../../../../../shared/components/utm/table/utm-table/utm-table-detail/utm-table-detail.component';
import {LOG_ANALYZER_TOTAL_ITEMS} from '../../../../../shared/constants/log-analyzer.constant';
import {ITEMS_PER_PAGE} from '../../../../../shared/constants/pagination.constants';
import {ElasticDataTypesEnum} from '../../../../../shared/enums/elastic-data-types.enum';
import {UtmFieldType} from '../../../../../shared/types/table/utm-field.type';

@Component({
  selector: 'app-alert-events-related',
  templateUrl: './alert-events-related.component.html',
  styleUrls: ['./alert-events-related.component.scss']
})
export class AlertEventsRelatedComponent implements OnInit {

  fields: UtmFieldType[]  = [
    {field: 'timestamp', label: '@timestamp', visible: true, type: ElasticDataTypesEnum.DATE},
  ];
  @Input() events: any[] = [];
  displayedLogs: any[] = [];
  page = 1;
  readonly totalItems = LOG_ANALYZER_TOTAL_ITEMS;
  itemsPerPage = ITEMS_PER_PAGE;
  readonly componentDetail = UtmTableDetailComponent;
  sortField = '';
  sortDirection: 'asc' | 'desc' = 'asc';

  constructor() { }

  ngOnInit() {
    this.applyFilters();
  }


  onPageChange($event: number) {
    this.page = $event;
    this.applyFilters();
  }

  onRemoveColumn($event: UtmFieldType) {

  }

  onSizeChange($event: number) {
    this.itemsPerPage = $event;
    this.applyFilters();
  }

  onSortBy($event: string) {

  }

  applyFilters() {
    const size = this.itemsPerPage;
    const filtered = [...this.events];

    // Sorting
    if (this.sortField) {
      filtered.sort((a, b) => {
        if (a[this.sortField] < b[this.sortField]) { return this.sortDirection === 'asc' ? -1 : 1; }
        if (a[this.sortField] > b[this.sortField]) { return this.sortDirection === 'asc' ? 1 : -1; }
        return 0;
      });
    }

    // Pagination
    const start = (this.page - 1) * size;
    this.displayedLogs = filtered.slice(start, start + size);
  }
}
