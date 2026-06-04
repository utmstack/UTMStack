import {Component, EventEmitter, Input, OnInit, Output, OnChanges, SimpleChanges, ViewChild, AfterViewInit} from '@angular/core';
import {Observable} from 'rxjs';
import {DataType} from '../../../rule-management/models/rule.model';
import {DataTypeService} from '../../../rule-management/services/data-type.service';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {LogstashService} from '../../../shared/services/logstash/logstash.service';
import {LogstashFilterType} from '../../shared/types/logstash-filter.type';
import {UtmPipeline} from '../../shared/types/logstash-stats.type';

@Component({
  selector: 'app-logstash-filter-create',
  templateUrl: './logstash-filter-create.component.html',
  styleUrls: ['./logstash-filter-create.component.scss']
})
export class LogstashFilterCreateComponent implements OnInit, OnChanges {
  @Output() filterEditClose = new EventEmitter<LogstashFilterType>();
  @Output() close = new EventEmitter<string>();
  @Input() filter: LogstashFilterType;
  @Input() pipeline: UtmPipeline;
  @Input() dataType: DataType;
  @Input() disable = false;
  @ViewChild('editorContainer') editorContainer: any;

  types$: Observable<DataType[]>;
  daTypeRequest: {page: number, size: number} = {
    page: -1,
    size: 10
  };

  saving = false;
  show = true;
  error = false;
  loadingDataTypes = false;
  private editorInstance: any;

  editorOptions: any = {
    theme: 'vs-light',
    language: 'yaml',
    automaticLayout: true,
    fontSize: 13,
    fontFamily: 'Courier New, Monaco, Menlo, monospace',
    lineNumbers: 'on',
    wordWrap: 'on',
    minimap: { enabled: false },
    scrollBeyondLastLine: false,
    formatOnPaste: true,
    formatOnType: true,
    tabSize: 2,
    insertSpaces: true,
    readOnly: false
  };

  constructor(private logstashFilterService: LogstashService,
              private utmToastService: UtmToastService,
              private dataTypeService: DataTypeService ) {
  }

  ngOnInit() {
    if (!this.filter || !this.filter.id) {
      this.filter = { logstashFilter: '', filterName: '', datatype: this.dataType};
    }

    this.types$ = this.dataTypeService.type$;
    this.loadDataTypes();
    this.updateEditorOptions();
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (changes && changes['disable']) {
      this.updateEditorOptions();
    }
  }

  private forceEditorLayout(): void {
    try {
      if (this.editorContainer) {
        const editorInstance = (this.editorContainer as any).editor;
        if (editorInstance && editorInstance.layout) {
          editorInstance.layout();
        }
      }
    } catch (e) {
      console.warn('Error al forzar layout del editor Monaco:', e);
    }
  }

  private updateEditorOptions(): void {
    this.editorOptions = {
      ...this.editorOptions,
      readOnly: this.disable
    };
  }


  save() {
    this.saving = true;
    if (this.filter.id) {
      this.updateFilter();
    } else {
      this.createFilter();
    }
  }

  updateFilter() {
    this.logstashFilterService.update(this.filter).subscribe(response => {
      this.onSaveSuccess('Log filter updated successfully', response.body);
    }, error => {
      this.onSaveError();
    });
  }

  onSaveSuccess(text: string, response: any) {
    this.utmToastService.showSuccessBottom(text);
    this.saving = false;
    this.filterEditClose.emit(response);
  }

  onSaveError() {
    this.saving = false;
    this.utmToastService.showError('Error saving log filter', 'Error while trying to change log filter');
  }

  createFilter() {
    this.logstashFilterService.create(this.filter, {pipelineId: this.pipeline.id}).subscribe(response => {
      this.onSaveSuccess('Log filter created successfully', response.body);
    }, error => {
      this.onSaveError();
    });
  }

  loadDataTypes() {
    this.daTypeRequest.page = this.daTypeRequest.page + 1;
    this.loadingDataTypes = true;

    this.dataTypeService.getAll(this.daTypeRequest)
      .subscribe( data => {
        this.loadingDataTypes = false;
      });
  }
  trackByFn(type: DataType) {
    return type.id;
  }

  onEditorChange(value: string): void {
    this.filter.logstashFilter = value;
  }
}
