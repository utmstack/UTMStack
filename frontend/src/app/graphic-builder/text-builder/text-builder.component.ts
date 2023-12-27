import {Location} from '@angular/common';
import {ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {AngularEditorConfig} from '@kolkov/angular-editor';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {NgxSpinnerService} from 'ngx-spinner';
import {VisualizationType} from '../../shared/chart/types/visualization.type';
import {ChartTypeEnum} from '../../shared/enums/chart-type.enum';
import {DataNatureTypeEnum} from '../../shared/enums/nature-data.enum';
import {RouteCallbackEnum} from '../../shared/enums/route-callback.enum';
import {VisualizationService} from '../visualization/shared/services/visualization.service';
import {VisualizationSaveComponent} from '../visualization/visualization-save/visualization-save.component';
import {DashboardStatusEnum} from "../dashboard-builder/shared/enums/dashboard-status.enum";

@Component({
  selector: 'app-text-builder',
  templateUrl: './text-builder.component.html',
  styleUrls: ['./text-builder.component.scss']
})
export class TextBuilderComponent implements OnInit {
  visualization: VisualizationType;
  mode: string;
  htmlContent: any;
  config: AngularEditorConfig = {
    editable: true,
    spellcheck: true,
    height: 'auto',
    minHeight: '65vh',
    maxHeight: 'auto',
    width: 'auto',
    minWidth: '0',
    translate: 'yes',
    enableToolbar: true,
    showToolbar: true,
    placeholder: 'Enter text here...',
    defaultParagraphSeparator: '',
    defaultFontName: '',
    defaultFontSize: '',
    sanitize: true,
    toolbarPosition: 'top',
    fonts: [
      {class: 'Roboto', name: 'Roboto'},
      {class: 'arial', name: 'Arial'},
      {class: 'times-new-roman', name: 'Times New Roman'},
      {class: 'calibri', name: 'Calibri'},
      {class: 'comic-sans-ms', name: 'Comic Sans MS'}
    ],
  };

  callback: RouteCallbackEnum;
  chartTypeEnum = ChartTypeEnum;
  visualizationId: number;
  chart: ChartTypeEnum;
  type: DataNatureTypeEnum;
  private patternId: number;

  constructor(private spinner: NgxSpinnerService,
              private route: ActivatedRoute,
              private router: Router,
              private modalService: NgbModal,
              private cdr: ChangeDetectorRef,
              private visualizationService: VisualizationService,
              private location: Location) {
    route.queryParams.subscribe(params => {
      this.visualizationId = params.visualizationId;
      this.chart = params.chart;
      this.type = params.type;
      this.mode = params.mode;
      this.patternId = params.patternId;
      if (params.callback) {
        this.callback = params.callback;
      }
    });

  }

  ngOnInit() {
    if (this.mode === 'edit') {
      this.visualizationService.find(this.visualizationId).subscribe(vis => {
        this.visualization = vis.body;
      });
    } else {
      this.visualization = {
        aggregationType: null,
        chartConfig: '',
        chartAction: null,
        filterType: null,
        chartType: this.chart,
        eventType: this.type,
        userCreated: null,
        name: '',
        idPattern: this.patternId
      };
    }
  }

  saveVisualization() {
    const modal = this.modalService.open(VisualizationSaveComponent, {centered: true});
    modal.componentInstance.visualization = this.visualization;
    modal.componentInstance.callback = this.callback;
    modal.componentInstance.mode = this.mode;
  }

  cancel() {
    if (this.callback === RouteCallbackEnum.DASHBOARD) {
      this.router.navigate(['/creator/dashboard/builder'],
          {
            queryParams:
                {
                  mode: DashboardStatusEnum.DRAFT
                }
          });
    } else {
      this.location.back();
    }
  }
}
