import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {LOG_INDEX_WINLOGBEAT} from '../../../../../../shared/constants/main-index-pattern.constant';
import {ElasticDataTypesEnum} from '../../../../../../shared/enums/elastic-data-types.enum';
import {ElasticOperatorsEnum} from '../../../../../../shared/enums/elastic-operators.enum';
import {ElasticSearchIndexService} from '../../../../../../shared/services/elasticsearch/elasticsearch-index.service';
import {ElasticFilterType} from '../../../../../../shared/types/filter/elastic-filter.type';
import {UtmFieldType} from '../../../../../../shared/types/table/utm-field.type';
import {FileFiltersBehavior} from '../../../behavior/file-filters.behavior';

@Component({
    selector: 'app-file-generic-filter',
    templateUrl: './file-generic-filter.component.html',
    styleUrls: ['./file-generic-filter.component.scss']
})
export class FileGenericFilterComponent implements OnInit {
    @Output() filterGenericChange = new EventEmitter<ElasticFilterType>();
    @Input() fieldFilter: UtmFieldType;
    activeFilters: ElasticFilterType[] = [];
    fieldValues: object;
    loading = true;
    selected: any[] = [];
    search: string;
    searching = false;
    loadingMore = false;
    top = 6;
    filter: ElasticFilterType;
    sort: { orderByCount: boolean, sortAsc: boolean } = {orderByCount: true, sortAsc: false};

    constructor(private elasticSearchIndexService: ElasticSearchIndexService,
                private fileFiltersBehavior: FileFiltersBehavior) {
    }

    ngOnInit() {
        // this.getFieldValues();
        /**
         * If filter is tags subscribe to changes to reload data on add new tag on alert
         */
        /**
         * Reset all values of selected filter
         */
        this.fileFiltersBehavior.$fileResetFilter.subscribe(reset => {
            if (reset) {
                this.selected = [];
            }
        });
        this.fileFiltersBehavior.$fileDeleteFilterValue.subscribe(deleteFilter => {
            if (deleteFilter) {
                const deleteField = deleteFilter.field.replace('.keyword', '');
                if (this.fieldFilter.field === deleteField) {
                    const selectedIndex = this.selected.findIndex(value => value === deleteFilter.value);
                    if (selectedIndex !== -1) {
                        this.selected.splice(selectedIndex, 1);
                    }
                }
            }
        });
        this.fileFiltersBehavior.$fileFilters.subscribe(fileFilters => {
            if (fileFilters) {
                this.activeFilters = fileFilters;
                this.getFieldValues();
                const index = fileFilters.findIndex(value => value.field.replace('.keyword', '') === this.fieldFilter.field);
                if (index !== -1) {
                    let values: any[] = fileFilters[index].value;
                    if (typeof values === 'string') {
                        values = [values];
                    }
                    if (values && values.length > 0) {
                        for (const val of values) {
                            if (!this.selected.includes(val)) {
                                this.selected.push(val);
                            }
                        }
                    }
                }
            }
        });
    }

    getFieldValues() {
        const field = this.setFieldKeyword();
        const filters = this.activeFilters
            .filter(value => !value.field.includes(field));
        if (this.search !== undefined && this.search !== '') {
            filters.push({
                field: this.fieldFilter.field,
                operator: ElasticOperatorsEnum.CONTAIN,
                value: this.search
            });
        }
        const req = {
          // dataOrigin: DataNatureTypeEnum.EVENT,
          field,
          filters,
          index: LOG_INDEX_WINLOGBEAT,
          orderByCount: this.sort.orderByCount,
          sortAsc: this.sort.sortAsc,
          top: this.top
        };
        this.elasticSearchIndexService.getValuesWithCount(req).subscribe(response => {
            this.fieldValues = response.body;
            this.loading = false;
            this.searching = false;
            this.loadingMore = false;
        });
    }

    setFieldKeyword(): string {
        return this.fieldFilter.type === ElasticDataTypesEnum.STRING ? this.fieldFilter.field + '.keyword' : this.fieldFilter.field;
    }

    valueHasData(): boolean {
        return Object.keys(this.fieldValues).length > 0;
    }

    searchInValues($event: string) {
        this.search = $event;
        this.searching = true;
        this.getFieldValues();
    }

    selectValue(value: any) {
        const index = this.selected.findIndex(val => val === value);
        if (index === -1) {
            this.selected.push(value);
        } else {
            this.selected.splice(index, 1);
        }
        this.emitCurrentFilter();
    }

    emitCurrentFilter() {
        const field = this.fieldFilter.type === ElasticDataTypesEnum.STRING ? this.fieldFilter.field + '.keyword' : this.fieldFilter.field;
        this.filter = {
            value: this.selected,
            operator: ElasticOperatorsEnum.IS_ONE_OF,
            field
        };
        this.filterGenericChange.emit(this.filter);
    }

    onScroll() {
        this.top += 10;
        this.loadingMore = true;
        this.getFieldValues();
    }

    onSortValuesChange($event: { orderByCount: boolean; sortAsc: boolean }) {
        this.sort = $event;
        this.getFieldValues();
    }
}
