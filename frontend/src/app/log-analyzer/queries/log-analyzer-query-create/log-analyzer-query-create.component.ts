import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {ElasticFilterType} from '../../../shared/types/filter/elastic-filter.type';
import {UtmIndexPattern} from '../../../shared/types/index-pattern/utm-index-pattern';
import {UtmFieldType} from '../../../shared/types/table/utm-field.type';
import {InputClassResolve} from '../../../shared/util/input-class-resolve';
import {LogAnalyzerQueryService} from '../../shared/services/log-analyzer-query.service';
import {LogAnalyzerQueryType} from '../../shared/type/log-analyzer-query.type';

@Component({
  selector: 'app-log-analyzer-query-create',
  templateUrl: './log-analyzer-query-create.component.html',
  styleUrls: ['./log-analyzer-query-create.component.scss']
})
export class LogAnalyzerQueryCreateComponent implements OnInit {
  queryForm: FormGroup;
  @Input() columns: UtmFieldType[];
  @Input() filters: ElasticFilterType[];
  @Input() pattern: UtmIndexPattern;
  @Output() querySaved = new EventEmitter<string>();
  @Input() query: LogAnalyzerQueryType;
  creating = false;
  private saveMode: boolean;

  constructor(public activeModal: NgbActiveModal,
              public inputClassResolve: InputClassResolve,
              private fb: FormBuilder,
              private utmToastService: UtmToastService,
              private analyzerQueryService: LogAnalyzerQueryService) {
  }

  ngOnInit() {
    this.initFormSaveVis();
    if (this.query) {
      this.queryForm.patchValue(this.query);
    }
    this.queryForm.get('filtersType').setValue(this.filters);
    this.queryForm.get('columnsType').setValue(this.columns);
    this.queryForm.get('idPattern').setValue(this.pattern.id);
  }

  initFormSaveVis() {
    this.queryForm = this.fb.group(
      {
        name: ['', Validators.required],
        description: [''],
        filtersType: [],
        columnsType: [],
        id: [],
        owner: [],
        idPattern: [],
        creationDate: []
      });
  }

  saveAsNew($event: boolean) {
    this.saveMode = $event;
    if ($event) {
      this.queryForm.get('name').setValue(this.query.name + '-Clone');
      this.queryForm.get('id').setValue(null);
    } else {
      this.queryForm.get('name').setValue(this.query.name);
    }
  }

  saveQuery() {
    this.creating = true;
    if (this.query && !this.saveMode) {
      this.editQuery();
    } else {
      this.createQuery();
    }
  }

  createQuery() {
    this.analyzerQueryService.create(this.queryForm.value).subscribe(indexPattern => {
      this.utmToastService.showSuccessBottom('Query saved successfully');
      this.activeModal.close();
      this.querySaved.emit('created');
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error', 'Error saving query');
    });
  }

  editQuery() {
    this.analyzerQueryService.update(this.queryForm.value).subscribe(indexPattern => {
      this.utmToastService.showSuccessBottom('Query updated successfully');
      this.activeModal.close();
      this.querySaved.emit('saved');
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error', 'Error updating query');
    });
  }
}
