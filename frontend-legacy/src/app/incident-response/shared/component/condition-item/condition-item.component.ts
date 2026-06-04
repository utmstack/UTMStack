import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormGroup} from '@angular/forms';
import {Observable} from 'rxjs';
import {map, tap} from 'rxjs/operators';
import {
  OperatorService
} from '../../../../shared/components/utm/filters/utm-elastic-filter/shared/util/operator.service';
import {ALERT_INDEX_PATTERN} from '../../../../shared/constants/main-index-pattern.constant';
import {ElasticDataTypesEnum} from '../../../../shared/enums/elastic-data-types.enum';
import {ElasticOperatorsEnum} from '../../../../shared/enums/elastic-operators.enum';
import {ElasticSearchIndexService} from '../../../../shared/services/elasticsearch/elasticsearch-index.service';
import {ElasticSearchFieldInfoType} from '../../../../shared/types/elasticsearch/elastic-search-field-info.type';

@Component({
  selector: 'app-condition-item',
  templateUrl: './condition-item.component.html',
  styleUrls: ['./condition-item.component.scss']
})
export class ConditionItemComponent implements OnInit {
  @Input() group: FormGroup;
  @Input() index: number;
  @Input() fields: ElasticSearchFieldInfoType[];
  @Output() delete: EventEmitter<number> = new EventEmitter();
  filteredFields: ElasticSearchFieldInfoType[] = [];
  protected readonly operatorEnum = ElasticOperatorsEnum;
  selectableOperators = [ElasticOperatorsEnum.IS_ONE_OF,
    ElasticOperatorsEnum.IS_NOT_ONE_OF,
    ElasticOperatorsEnum.CONTAIN_ONE_OF,
    ElasticOperatorsEnum.DOES_NOT_CONTAIN_ONE_OF];
  operators = [];
  req = {
    page: 0,
    size: 10,
    indexPattern: ALERT_INDEX_PATTERN,
    keyword: ''
  };
  values$: Observable<string[]>;
  loading = false;
  multiple = true;

  constructor(private elasticSearchIndexService: ElasticSearchIndexService,
              private operatorsService: OperatorService) { }

  ngOnInit() {
    this.filteredFields = this.fields.filter(field => field.name && !field.name.includes('.keyword'));
  }

  onScroll() {
    this.req = {
      ...this.req,
      size: this.req.size + 10
    };

    this.getValues();
  }

  getValuesForField(key: string | null) {

    this.req = {
      ...this.req,
      keyword: key ? key + '.keyword' : this.req.keyword
    };
    this.refreshOperators();
    this.getValues();
  }

  getValues() {
    this.loading = true;
    this.values$ = this.elasticSearchIndexService.getElasticFieldValues(this.req)
      .pipe(
        tap((res) => this.loading = !this.loading),
        map(res => res.body));
  }

  refreshOperators() {
    this.operators = this.operatorsService.getOperators(this.field, this.operators);
  }

  /**
   * Return boolean, if show input or select values
   */
  applySelectFilter(): boolean {
    if (this.field) {
      if (this.field.type === ElasticDataTypesEnum.TEXT ||
        this.field.type === ElasticDataTypesEnum.STRING) {
        if (this.field.name.includes('.keyword')) {
          return this.operatorFieldSelectable();
        } else {
          // if type of current filter is not keyword return result of validation if current operator is selectable or not
          return this.selectableOperators.includes(this.operator);
        }
      } else if (this.field.type === ElasticDataTypesEnum.DATE) {
        return this.operator === ElasticOperatorsEnum.IS_ONE_OF ||
          this.operator === ElasticOperatorsEnum.IS_NOT_ONE_OF;
      } else {
        // if current field is not a date or text return result of function if field filter value cant show select or input
        return this.operatorFieldSelectable();
      }
    } else {
      return false;
    }
  }

  operatorFieldSelectable(): boolean {
    return this.operator === ElasticOperatorsEnum.IS || this.operator === ElasticOperatorsEnum.IS_NOT ||
      this.operator === ElasticOperatorsEnum.IS_ONE_OF || this.operator === ElasticOperatorsEnum.IS_NOT_ONE_OF ||
      this.operator === ElasticOperatorsEnum.CONTAIN_ONE_OF || this.operator === ElasticOperatorsEnum.DOES_NOT_CONTAIN_ONE_OF;
  }


  get field(): ElasticSearchFieldInfoType {
    const field = this.group.get('field').value;
    const index = this.fields.findIndex(value => value.name === field);
    if (index !== -1) {
      return this.fields[index];
    }
  }

  get operator() {
    return this.group.get('operator').value;
  }


  selectOperator($event: MouseEvent) {

  }
  isMultipleSelectValue() {
    this.multiple = this.group.get('operator').value === ElasticOperatorsEnum.IS_ONE_OF ||
      this.group.get('operator').value === ElasticOperatorsEnum.IS_NOT_ONE_OF ||
      this.group.get('operator').value === ElasticOperatorsEnum.CONTAIN_ONE_OF ||
      this.group.get('operator').value === ElasticOperatorsEnum.DOES_NOT_CONTAIN_ONE_OF;
  }

  onOperatorChange($event: {}) {
    this.group.get('value').reset();
    this.isMultipleSelectValue();
  }

  onSearch(term: { term: string }) {
    this.filteredFields = this.fields.filter(field => field.name && !field.name.includes('.keyword'));
    if (!term.term) {
      return;
    }

    const searchTerm = term.term.toLowerCase();
    this.filteredFields = this.filteredFields
      .filter(field => field.name.toLowerCase().includes(searchTerm))
      .sort((a, b) => {
        const aStarts = a.name.toLowerCase().startsWith(searchTerm);
        const bStarts = b.name.toLowerCase().startsWith(searchTerm);

        if (aStarts && !bStarts) { return -1; }
        if (!aStarts && bStarts) { return 1; }
        return a.name.localeCompare(b.name);
      });
  }
}
