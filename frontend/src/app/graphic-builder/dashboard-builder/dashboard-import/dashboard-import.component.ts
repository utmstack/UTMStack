import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {UtmDashboardVisualizationType} from '../../../shared/chart/types/dashboard/utm-dashboard-visualization.type';
import {UtmDashboardType} from '../../../shared/chart/types/dashboard/utm-dashboard.type';
import {VisualizationType} from '../../../shared/chart/types/visualization.type';
import {UTM_CHART_ICONS} from '../../../shared/constants/icons-chart.const';
import {ChartTypeEnum} from '../../../shared/enums/chart-type.enum';
import {ElasticDataTypesEnum} from '../../../shared/enums/elastic-data-types.enum';
import {UtmDashboardService} from '../shared/services/utm-dashboard.service';

@Component({
  selector: 'app-dashboard-import',
  templateUrl: './dashboard-import.component.html',
  styleUrls: ['./dashboard-import.component.scss']
})
export class DashboardImportComponent implements OnInit {
  step = 1;
  stepCompleted: number[] = [];
  importing: any;
  dashImport: UtmDashboardVisualizationType[] = [];
  dashboards: UtmDashboardType[] = [];
  totalItems: number;
  page = 1;
  itemsPerPage = 5;
  pageStart = 0;
  pageEnd = 5;
  detailOf: number;
  visualizations: VisualizationType[] = [];
  @Output() dashboardImported = new EventEmitter<string>();
  override = false;

  constructor(private dashboardService: UtmDashboardService,
              private activeModal: NgbActiveModal,
              private utmToastService: UtmToastService) {
  }

  ngOnInit() {
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

  onFileImportLoad($event: any[]) {
    this.dashImport = $event;
    this.loadListDashboard($event);
  }

  loadListDashboard(visualizationDashboard: UtmDashboardVisualizationType[]) {
    for (const visDash of visualizationDashboard) {
      if (this.dashboards.findIndex(value => value.id === visDash.dashboard.id) === -1) {
        this.dashboards.push(visDash.dashboard);
      }
    }
    this.totalItems = this.dashboards.length;
  }

  import() {
    this.importing = true;
    this.convertPropertiesToString().then(visDash => {
      this.dashboardService.import({
        dashboards: visDash,
        override: this.override
      }).subscribe(response => {
        this.importing = false;
        this.activeModal.close();
        this.utmToastService.showSuccessBottom('Dashboards imported successfully');
        this.dashboardImported.emit('success');
      }, error => {
        this.importing = false;
        this.utmToastService.showError('Error', 'Error occurring while trying to import dashboards');
      });
    });
  }

  convertPropertiesToString(): Promise<UtmDashboardVisualizationType[]> {
    return new Promise<UtmDashboardVisualizationType[]>(resolve => {
      this.dashImport.forEach(value => {
        if (typeof value.visualization.chartConfig !== ElasticDataTypesEnum.STRING) {
          value.visualization.chartConfig = JSON.stringify(value.visualization.chartConfig);
        }
        if (typeof value.visualization.chartAction !== ElasticDataTypesEnum.STRING) {
          value.visualization.chartAction = JSON.stringify(value.visualization.chartAction);
        }
      });
      resolve(this.dashImport);
    });
  }

  loadPage(page: number) {
    this.pageEnd = page * this.itemsPerPage;
    this.pageStart = this.pageEnd - this.itemsPerPage;
  }

  viewVisualizationOf(dash: UtmDashboardType) {
    this.visualizations = [];
    if (this.detailOf === dash.id) {
      this.detailOf = 0;
    } else {
      this.detailOf = dash.id;
      const tempDash = this.dashImport.filter(value => value.dashboard.id === dash.id);
      for (const dashVis of tempDash) {
        this.visualizations.push(dashVis.visualization);
      }
    }
  }

  chartIconResolver(chartType: ChartTypeEnum) {
    return UTM_CHART_ICONS[chartType];
  }

  deleteDashboard(dash: UtmDashboardType) {
    this.dashImport.forEach((value, index) => {
      if (value.idDashboard === dash.id) {
        this.dashImport.splice(index, 1);
      }
    });
    const indexOfDelete = this.dashboards.findIndex(value => value.id === dash.id);
    this.dashboards.splice(indexOfDelete, 1);
  }

  deleteVisualization(vis: VisualizationType, dash: UtmDashboardType) {
    const indexDashboard = this.dashImport.findIndex(value => value.id === dash.id);
    this.dashImport.splice(indexDashboard, 1);
    const indexVis = this.visualizations.findIndex(value => value.id === vis.id);
    this.visualizations.splice(indexVis, 1);
  }
}
