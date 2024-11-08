import {Component, EventEmitter, Input, OnDestroy, OnInit, Output} from '@angular/core';
import {Subject} from 'rxjs';
import {filter, takeUntil} from 'rxjs/operators';
import {
  UtmFilterBehavior
} from '../../../shared/components/utm/filters/utm-elastic-filter/shared/behavior/utm-filter.behavior';
import {ElasticDataTypesEnum} from '../../../shared/enums/elastic-data-types.enum';
import {ElasticOperatorsEnum} from '../../../shared/enums/elastic-operators.enum';
import {FieldDataService} from '../../../shared/services/elasticsearch/field-data.service';
import {ElasticSearchFieldInfoType} from '../../../shared/types/elasticsearch/elastic-search-field-info.type';
import {ElasticFilterType} from '../../../shared/types/filter/elastic-filter.type';
import {IndexFieldController} from '../../shared/behaviors/index-field-controller.behavior';
import {IndexPatternBehavior} from '../../shared/behaviors/index-pattern.behavior';
import {LogAnalyzerDataChangeType} from '../../shared/type/log-analyzer-data-change.type';

@Component({
  selector: 'app-log-analyzer-field',
  templateUrl: './log-analyzer-field.component.html',
  styleUrls: ['./log-analyzer-field.component.scss']
})
export class LogAnalyzerFieldComponent implements OnInit, OnDestroy {
  @Input() fieldSelected: ElasticSearchFieldInfoType[];
  @Input() uuid: string;
  @Input() filters: ElasticFilterType[];
  @Output() columnChange = new EventEmitter<ElasticSearchFieldInfoType[]>();
  fields: ElasticSearchFieldInfoType[] = [];
  fieldsOriginal: ElasticSearchFieldInfoType[] = [];
  loadingFields = false;
  searching = false;
  pattern: string;
  pageStart = 0;
  pageEnd = 100;
  destroy$ = new Subject<void>();

  constructor(private indexPatternBehavior: IndexPatternBehavior,
              private fieldDataService: FieldDataService,
              private utmFilterBehavior: UtmFilterBehavior,
              private indexFieldController: IndexFieldController) {
  }

  ngOnInit() {
    this.fieldSelected = this.fieldSelected ? this.fieldSelected : [];
    this.indexPatternBehavior.$pattern
      .pipe(
        takeUntil(this.destroy$),
        filter((dataChange: LogAnalyzerDataChangeType) => {
          console.log('UUID', this.uuid);
          console.log('dataChange', !!dataChange);
          console.log('condition', !!dataChange && this.uuid === dataChange.tabUUID);
          return !!dataChange && this.uuid === dataChange.tabUUID;
        }))
      .subscribe((dataChange: LogAnalyzerDataChangeType) => {
        this.pageStart = 0;
        this.pageEnd = 100;
        this.pattern = dataChange.pattern.pattern;
        this.fieldDataService.getFields(this.pattern).subscribe(field => {
          this.fields = field;
          this.fieldsOriginal = field;
          this.fieldSelected = this.fieldSelected ? this.fieldSelected : [{
            name: '@timestamp',
            type: ElasticDataTypesEnum.DATE
          }];
          this.columnChange.emit(this.fieldSelected);
          this.loadingFields = false;
        }, error => {
          this.loadingFields = false;
          this.fields = [];
        });
      });

    this.indexFieldController.$field.subscribe(field => {
      if (field) {
        const indexToAdd = this.getIndexField(field);
        if (indexToAdd !== -1) {
          this.addToColumns(this.fieldsOriginal[indexToAdd]);
        } else {
          const regex = /\.(\d+)\./g;
          if (regex.test(field)) {
            const referenceField = field.replace(regex, '.');
            const refFieldIndex = this.getIndexField(referenceField);
            if (refFieldIndex !== -1) {
              this.addToColumns({name: field, type: this.fieldsOriginal[refFieldIndex].type});
            }
          }
        }
      }
    });
  }

  getIndexField(field: string): number {
    return this.fieldsOriginal.findIndex(value => value.name === field);
  }

  onSearch($event: string) {
    this.searching = true;
    this.doSearch($event).then(() => this.searching = false);
  }

  doSearch(search: string): Promise<boolean> {
    return new Promise<boolean>(resolve => {
      if (search !== '') {
        this.fields = this.fieldsOriginal.filter(value => {
          return value.name.toLowerCase().includes(search.toLowerCase()) && !value.name.includes('.keyword');
        });
      } else {
        this.fields = this.fieldsOriginal.filter(value => {
          return !value.name.includes('.keyword');
        });
      }
      resolve(true);
    });
  }

  addToColumns(field: ElasticSearchFieldInfoType) {
    const index = this.fieldSelected.findIndex(value => value.name === field.name);
    if (index === -1) {
      if (field.type !== ElasticDataTypesEnum.TEXT) {
        this.fieldSelected.push(field);
        this.utmFilterBehavior.$filterExistChange
          .next({field: field.name, operator: ElasticOperatorsEnum.EXIST});
      } else {
        const indexFieldKeyword = this.fieldsOriginal.findIndex(fie => fie.name === field.name + '.keyword');
        if (indexFieldKeyword !== -1) {
          this.utmFilterBehavior.$filterExistChange
            .next({field: this.fieldsOriginal[indexFieldKeyword].name, operator: ElasticOperatorsEnum.EXIST});
          this.fieldSelected.push(this.fieldsOriginal[indexFieldKeyword]);
        } else {
          this.utmFilterBehavior.$filterExistChange
            .next({field: field.name, operator: ElasticOperatorsEnum.EXIST});
          this.fieldSelected.push(field);
        }
      }
    } else {
      this.fieldSelected.splice(index, 1);
    }
    this.columnChange.emit(this.fieldSelected);
  }

  isInColumns(field: ElasticSearchFieldInfoType): boolean {
    const fieldName = field.type === ElasticDataTypesEnum.TEXT ? field.name + '.keyword' : field.name;
    return this.fieldSelected.findIndex(value => value.name === fieldName) !== -1;
  }

  onScroll() {
    this.pageEnd += 50;
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
