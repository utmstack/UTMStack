import {
  AfterViewInit,
  Component, ElementRef,
  EventEmitter,
  Input, OnChanges,
  OnDestroy,
  OnInit,
  Output, QueryList, SimpleChanges, Type,
  ViewChild, ViewChildren,
  ViewContainerRef
} from '@angular/core';
import {container} from '@angular/core/src/render3';
import {Subject} from 'rxjs';
import {IndexPatternBehavior} from '../../../../../../log-analyzer/shared/behaviors/index-pattern.behavior';
import {SortEvent} from '../../../../../directives/sortable/type/sort-event';
import {ElasticDataTypesEnum} from '../../../../../enums/elastic-data-types.enum';
import {UtmDateFormatEnum} from '../../../../../enums/utm-date-format.enum';
import {UtmFieldType} from '../../../../../types/table/utm-field.type';
import {
  convertObjectToKeyValueArray, extractFieldValueFromKvArray,
  extractValueFromObjectByPath
} from '../../../../../util/get-value-object-from-property-path.util';
import {SUMMARY_COLUMNS} from './summary-fields';

@Component({
  selector: 'app-dynamic-table',
  templateUrl: './dynamic-table.component.html',
  styleUrls: ['./dynamic-table.component.scss']
})
export class UtmDynamicTableComponent implements OnInit, OnChanges, OnDestroy {

  constructor(private indexPatternBehavior: IndexPatternBehavior) {
  }
  /**
   * Field to show in header
   */
  @Input() fields: UtmFieldType[] = [];

  /**
   * Data to show
   */
  @Input() rows: any[] = [];
  /**
   * Loading signal
   */
  @Input() loading: boolean;
  @Input() totalItems: number;
  @Input() page: number;
  @Input() itemsPerPage: number;
  /**
   * Show invisible columns in columns management
   */
  @Input() showInactiveColumns;
  /**
   * Determine if show detail or not default true
   */
  @Input() showDetail = true;
  /**
   * Determine if show delete in header
   */
  @Input() viewDeleteAction: boolean;
  /**
   * Component to show in detail
   */
  @Input() componentDetail: Type<any>;
  /**
   * Event emit when page change, number of page
   */
  @Output() pageChange = new EventEmitter<number>();
  /**
   * Event emit when size per page change
   */
  @Output() sizeChange = new EventEmitter<number>();
  /**
   * Event emit when sort table
   */
  @Output() sortBy = new EventEmitter<string>();
  /**
   * Event emit when columns was deleted, emit type UtmTableFieldType
   */
  @Output() removeColumn = new EventEmitter<UtmFieldType>();
  /**
   * Var to put table max width
   */
  @Input() tableWidth: number;
  /**
   * Determine if show edit columns property
   */
  @Input() pageable = true;

  @Input() initialExpandedId?: number;

  @Input() editableColumn = true;
  @ViewChild('container', {read: ViewContainerRef}) entry: ViewContainerRef;
  @ViewChild('container') containerRef!: ElementRef<HTMLDivElement>;
  viewDetail = -1;
  dataTypeEnum = ElasticDataTypesEnum;
  utmDateFormat = UtmDateFormatEnum;
  destroy$: Subject<void> = new Subject<void>();

  summaryColumns = SUMMARY_COLUMNS;

  @ViewChildren('summaryWrapper') summaryWrappers: QueryList<ElementRef>;

  ngOnInit(): void {
    if (this.initialExpandedId !== undefined) {
      this.viewDetail = this.initialExpandedId;
    }
  }

  onSort($event: SortEvent) {
    this.sortBy.emit($event.column + ',' + $event.direction);
  }

  resolveTdValue(row: any, td: UtmFieldType): any {
    const value = extractFieldValueFromKvArray(row, td);
    return td.type === this.dataTypeEnum.DATE && value === '-' ?
      null : value;
  }

  processSourceDocument(row: any): Promise<string> {
    return new Promise<string>(resolve => {
      let html = '';
      this.buildIterateFieldSource(row).then(content => {
        html = '<div class="truncate-by-height">' +
          '<span>' + '<dl class="source truncate-by-height">' +
          content
          + '</dl></span></div>';
      });
      resolve(html);
    });
  }

  buildIterateFieldSource(row): Promise<string> {
    return new Promise<string>(resolve => {
      let html = '';
      for (const source of convertObjectToKeyValueArray(row)) {
        html += '<dt>' + source.key + '</dt>' +
          '<dd><span>' + source.value + '</span>' +
          '</dd>';
      }
      resolve(html);
    });
  }

  /**
   * Find in all field if column type text has keyword
   * @param column Column to check if column cant be sort
   */
  isSortableColumn(column: UtmFieldType): boolean {
    if (column.type === ElasticDataTypesEnum.TEXT || column.type === ElasticDataTypesEnum.STRING) {
      return column.field.includes('.keyword');
    } else {
      return true;
    }
  }

  resolveTableHeaderName(column: UtmFieldType): string {
    if (column.label) {
      return column.label;
    } else {
      return column.field.includes('.keyword') ? column.field.replace('.keyword', '') : column.field;
    }
  }

  getSummaryChunks(row: any): { visible: any[], hidden: any[] } {
    if (!this.summaryColumns) { return { visible: [], hidden: [] }; }

    const all = this.summaryColumns.filter(col => this.resolveTdValue(row, col.key) !== '-');
    const container = this.containerRef.nativeElement;
    if (!container) { return { visible: all, hidden: [] }; }

    const visible: any[] = [];
    const tempDiv = document.createElement('div');
    tempDiv.style.position = 'absolute';
    tempDiv.style.visibility = 'hidden';
    tempDiv.style.display = 'flex';
    tempDiv.style.flexWrap = 'wrap';
    container.appendChild(tempDiv);

    let totalWidth = 0;
    for (const item of all) {
      const chip = document.createElement('div');
      chip.className = 'chip-summary';
      chip.innerText = `${item.title}: ${this.resolveTdValue(row, item.key)}`;
      tempDiv.appendChild(chip);

      const width = chip.getBoundingClientRect().width;
      totalWidth += width + 8; // padding + gap
      if (totalWidth < container.clientWidth) {
        visible.push(item);
      } else {
        break;
      }
    }

    container.removeChild(tempDiv);

    return {
      visible,
      hidden: all.slice(visible.length)
    };
  }


  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (!this.loading) {
      setTimeout(() => {
        this.summaryWrappers.forEach((wrapper, index) => {
          const el = wrapper.nativeElement;
          const row = this.rows[index];
          row._hasOverflow = el.scrollHeight > 30;
        });
      }, 300);
    }
  }

}
