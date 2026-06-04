import {AfterContentChecked, ChangeDetectorRef, Component, ElementRef, EventEmitter, OnInit, Output, ViewChild} from '@angular/core';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {VisualizationType} from '../../../shared/chart/types/visualization.type';
import {UTM_CHART_ICONS} from '../../../shared/constants/icons-chart.const';
import {IndexPatternService} from '../../../shared/services/elasticsearch/index-pattern.service';
import {UtmIndexPattern} from '../../../shared/types/index-pattern/utm-index-pattern';
import {VisualizationService} from '../shared/services/visualization.service';

@Component({
  selector: 'app-visualization-import',
  templateUrl: './visualization-import.component.html',
  styleUrls: ['./visualization-import.component.scss']
})
export class VisualizationImportComponent implements OnInit, AfterContentChecked {
  @Output() visualizationImported = new EventEmitter<string>();
  step = 1;
  stepCompleted: number[] = [];
  creating = false;
  visImport: VisualizationType[] = [];
  visualizations: VisualizationType[] = [];
  imported = 0;
  totalItems: number;
  page = 1;
  itemsPerPage = 5;
  pageStart = 0;
  pageEnd = 5;
  patterns: UtmIndexPattern[] = [];
  @ViewChild('content') content: ElementRef;
  visResolve: VisualizationType;
  importing = false;
  validNames: boolean;
  validPatterns: boolean;
  override = false;

  constructor(private indexPatternService: IndexPatternService,
              private visualizationService: VisualizationService,
              private modalService: NgbModal,
              private activeModal: NgbActiveModal,
              private utmToast: UtmToastService,
              private cdr: ChangeDetectorRef) {
  }

  ngOnInit() {
    this.loadIndexPatterns();
    this.loadVisualizations();
  }

  ngAfterContentChecked(): void {
    this.cdr.detectChanges();
  }

  loadIndexPatterns() {
    const req = {
      size: 10000,
      page: 0
    };
    this.indexPatternService.query(req).subscribe(index => {
      this.patterns = index.body;
    });
  }

  loadVisualizations() {
    const req = {
      size: 10000,
      page: 0
    };
    this.visualizationService.query(req).subscribe(vis => {
      this.visualizations = vis.body;
    });
  }

  chartIconResolver(chartType: string) {
    return UTM_CHART_ICONS[chartType];
  }

  nextStep() {
    this.stepCompleted.push(this.step);
    this.step += 1;
  }

  isCompleted(step: number) {
    return this.stepCompleted.findIndex(value => value === step) !== -1;
  }

  backStep() {
    this.stepCompleted.pop();
    this.step -= 1;
  }

  import() {
    this.importing = true;
    this.uploadVisualization().then(value => {
      this.importing = false;
      this.utmToast.showSuccessBottom('Visualizations imported successfully');
      this.visualizationImported.emit('success');
      this.activeModal.close();
    });
  }

  uploadVisualization(): Promise<boolean> {
    return new Promise<boolean>((resolve) => {
      this.visImport.forEach(vis => delete vis.id);
      const interval = setInterval(() => {
        if (this.imported < 100) {
          this.imported += 1;
        }
      }, 500);
      this.visualizationService.bulkImport({visualizations: this.visImport, override: this.override})
        .subscribe(imported => {
          resolve(true);
          this.imported = 100;
          clearInterval(interval);
        });
    });
  }


  onFileImportLoad($event) {
    this.visImport = $event;
    this.totalItems = this.visImport.length;
  }

  loadPage(page: number) {
    this.pageEnd = page * this.itemsPerPage;
    this.pageStart = this.pageEnd - this.itemsPerPage;
  }

  deleteVisualization(index: number) {
    this.visImport.splice(index, 1);
    this.totalItems = this.visImport.length;
  }

  haveNameProblems(vis: VisualizationType) {
    const valid = this.visualizations.findIndex(value => value.name === vis.name) !== -1;
    this.validNames = valid;
    return valid;
  }

  havePatternProblems(visResolve: VisualizationType) {
    // const valid = this.patterns.findIndex(value => value.id === visResolve.idPattern) === -1;
    // this.validPatterns = valid;
    return false;
  }

  changeVisName(vis: VisualizationType, name: string) {
    const indexVis = this.visImport.findIndex(value => value.id === vis.id);
    this.visImport[indexVis].name = name;
  }

  openResolverModal(vis: VisualizationType) {
    this.visResolve = vis;
    this.modalService.open(this.content, {size: 'sm', centered: true});
  }

  setPattern($event: any) {
    const indexVis = this.visImport.findIndex(value => value.name === this.visResolve.name);
    this.visImport[indexVis].pattern = $event;
    this.visImport[indexVis].idPattern = $event.id;
  }
}
