import {HttpErrorResponse, HttpResponse} from '@angular/common/http';
import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {filterFieldDataType} from '../../../../../graphic-builder/shared/util/elasticsearch/elastic-field.util';
import {UtmToastService} from '../../../../alert/utm-toast.service';
import {ITEMS_PER_PAGE} from '../../../../constants/pagination.constants';
import {ElasticDataTypesEnum} from '../../../../enums/elastic-data-types.enum';
import {ElasticSearchIndexService} from '../../../../services/elasticsearch/elasticsearch-index.service';
import {IndexPatternService} from '../../../../services/elasticsearch/index-pattern.service';
import {ElasticSearchFieldInfoType} from '../../../../types/elasticsearch/elastic-search-field-info.type';
import {ElasticsearchIndexInfoType} from '../../../../types/elasticsearch/elasticsearch-index-info.type';
import {UtmIndexPattern} from '../../../../types/index-pattern/utm-index-pattern';

@Component({
  selector: 'app-index-pattern-create',
  templateUrl: './index-pattern-create.component.html',
  styleUrls: ['./index-pattern-create.component.scss']
})
export class IndexPatternCreateComponent implements OnInit {
  @Output() indexPatternCreated = new EventEmitter<UtmIndexPattern>();
  @Input() indexPattern: UtmIndexPattern;
  indexes: ElasticsearchIndexInfoType[] = [];
  totalItems: any;
  page = 0;
  itemsPerPage = 5;
  loading = false;
  step = 1;
  stepCompleted: number[] = [];
  regex = '';
  fields: ElasticSearchFieldInfoType[] = [];
  loadingFields = false;
  showAdvanced = false;
  includeSystemIndex = false;
  creating = false;
  indexExist = false;
  patternLabel: string;
  id: number;

  constructor(public activeModal: NgbActiveModal,
              private indexPatternService: IndexPatternService,
              private elasticIndexService: ElasticSearchIndexService,
              private utmToastService: UtmToastService) {
  }

  ngOnInit() {
    if (this.indexPattern) {
      this.regex = this.indexPattern.pattern;
      this.id = this.indexPattern.id;
    }

    this.getIndexes();
  }


  getIndexes() {
    const req = {
      includeSystemIndex: this.includeSystemIndex,
      page: this.page,
      pattern: this.regex,
      size: this.itemsPerPage
    };
    this.elasticIndexService.getElasticIndex(req).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  searchMatchingRegex() {
    // this.regex = this.regex.replace('*', '') + '*';
    this.page = 0;
    this.getIndexes();
    this.getIndexPatterns();
  }

  loadPage(page: any) {
    this.page = page - 1;
    this.getIndexes();
  }

  backStep() {
    this.step -= 1;
  }

  nextStep() {
    this.stepCompleted.push(this.step);
    this.step += 1;
    this.getFieldsOfPattern();
  }

  isCompleted(step: number) {
    return this.stepCompleted.findIndex(value => value === step) !== -1;
  }

  showAdvancedSetting() {
    this.showAdvanced = !this.showAdvanced;
  }

  createIndexPattern() {
    this.stepCompleted.push(2);
    this.creating = true;
    if (this.id) {
      this.edit();
    } else {
      this.create();
    }
  }

  create() {
    const pattern = {
      pattern: this.regex,
      systemPattern: false
    };
    this.indexPatternService.create(pattern).subscribe(response => {
      this.success('created', response);
    }, error => {
      this.error('created', error);
    });
  }

  edit() {
    const pattern = {
      pattern: this.regex,
      systemPattern: false,
      id: this.id
    };
    this.indexPatternService.update(pattern).subscribe(response => {
      this.success('edited', response);
    }, error => {
      this.error('edited', error);
    });
  }

  success(action: 'edited' | 'created', response: HttpResponse<any>) {
    this.utmToastService.showSuccessBottom('Index pattern ' + action + ' successfully');
    this.activeModal.close();
    this.indexPatternCreated.emit(response.body);
  }

  error(action: 'edited' | 'created', error: HttpErrorResponse) {
    this.creating = false;
    this.utmToastService.showError('Error creating Index pattern',
      error.error.statusText);
  }

  onToggleChange($event: boolean) {
    this.loading = true;
    this.includeSystemIndex = $event;
    this.getIndexes();
  }

  getFieldsOfPattern() {
    this.loadingFields = true;
    const req = {
      indexPattern: this.regex
    };
    this.elasticIndexService.getElasticIndexField(req).subscribe(fields => {
      filterFieldDataType(fields.body, ElasticDataTypesEnum.DATE).subscribe(field => {
        this.loadingFields = false;
        this.fields = field;
      });
    });
  }

  getIndexPatterns() {
    const req = {
      page: 0,
      size: 1000,
      'pattern.equals': this.regex,
    };
    this.indexPatternService.query(req).subscribe(index => {
      this.indexExist = index.body.length > 0;
    });
  }

  private onSuccess(data, headers) {
    this.totalItems = headers.get('X-Total-Count');
    this.indexes = data;
    this.loading = false;
  }

  private onError(error) {
    // this.alertService.error(error.error, error.message, null);
  }

}
