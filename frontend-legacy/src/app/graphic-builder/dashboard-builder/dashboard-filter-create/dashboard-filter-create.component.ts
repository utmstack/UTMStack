import {HttpResponse} from '@angular/common/http';
import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UUID} from 'angular2-uuid';
import {resolveIcon} from 'src/app/shared/util/elastic-fields.util';
import {ElasticDataTypesEnum} from '../../../shared/enums/elastic-data-types.enum';
import {FieldDataService} from '../../../shared/services/elasticsearch/field-data.service';
import {IndexPatternService} from '../../../shared/services/elasticsearch/index-pattern.service';
import {ElasticSearchFieldInfoType} from '../../../shared/types/elasticsearch/elastic-search-field-info.type';
import {DashboardFilterType} from '../../../shared/types/filter/dashboard-filter.type';
import {UtmIndexPattern} from '../../../shared/types/index-pattern/utm-index-pattern';

@Component({
  selector: 'app-dashboard-filter-create',
  templateUrl: './dashboard-filter-create.component.html',
  styleUrls: ['./dashboard-filter-create.component.css']
})
export class DashboardFilterCreateComponent implements OnInit {
  @Input() filter: DashboardFilterType;
  @Output() filterAdd = new EventEmitter<DashboardFilterType>();
  filterForm: FormGroup;
  fields: ElasticSearchFieldInfoType[] = [];
  patterns: UtmIndexPattern[];

  constructor(public activeModal: NgbActiveModal,
              private fb: FormBuilder,
              private indexPatternService: IndexPatternService,
              private fieldDataBehavior: FieldDataService) {
  }

  ngOnInit() {
    this.getIndexPatterns();
    this.filterForm = this.fb.group({
      id: [UUID.UUID()],
      filterLabel: ['', Validators.required],
      indexPattern: ['', Validators.required],
      field: ['', Validators.required],
      multiple: [false],
      searchable: [true],
      clearable: [true],
      loadingText: ['Loading'],
      placeholder: [''],
      maxSelectedItems: [100]
    });
    if (this.filter) {
      this.getFields(this.filter.indexPattern);
      this.filterForm.setValue(this.filter);
    }
  }

  saveFilter() {
    this.filterAdd.emit(this.filterForm.value);
    this.activeModal.close();
  }

  getFields(indexPattern: string) {
    this.fieldDataBehavior.getFields(indexPattern).subscribe(field => {
      if (field) {
        this.fields = field.filter(value => value.type !== ElasticDataTypesEnum.TEXT || value.name.includes('.keyword'));
      }
    });
  }

  selectPattern($event: any) {
    this.filterForm.get('indexPattern').setValue($event.pattern);
    this.filterForm.get('field').setValue(null);
    this.fields = null;
    this.getFields($event.pattern);
  }

  resolveIcon(field: ElasticSearchFieldInfoType): string {
    return resolveIcon(field);
  }

  getIndexPatterns() {
    const req = {
      page: 0,
      size: 1000,
    };
    this.indexPatternService.query(req).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  private onError(error) {
    // this.alertService.error(error.error, error.message, null);
  }

  private onSuccess(data) {
    this.patterns = data;
  }

}
