import {HttpResponse} from '@angular/common/http';
import {Component, OnInit, QueryList, ViewChildren} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';

import * as moment from 'moment';
import {NgxSpinnerService} from 'ngx-spinner';
import {VisualizationType} from '../../../shared/chart/types/visualization.type';
import {UTM_CHART_ICONS} from '../../../shared/constants/icons-chart.const';
import {ITEMS_PER_PAGE} from '../../../shared/constants/pagination.constants';
import {SortableDirective} from '../../../shared/directives/sortable/sortable.directive';
import {SortEvent} from '../../../shared/directives/sortable/type/sort-event';
import {ActionInitParamsEnum, ActionInitParamsValueEnum} from '../../../shared/enums/action-init-params.enum';
import {ChartTypeEnum} from '../../../shared/enums/chart-type.enum';
import {SortByType} from '../../../shared/types/sort-by.type';
import {VisualizationQueryParamsEnum} from '../../shared/enums/visualization-query-params.enum';
import {VisualizationService} from '../shared/services/visualization.service';
import {VisualizationCreateComponent} from '../visualization-create/visualization-create.component';
import {VisualizationDeleteComponent} from '../visualization-delete/visualization-delete.component';
import {VisualizationImportComponent} from '../visualization-import/visualization-import.component';

@Component({
  selector: 'app-visualization-list',
  templateUrl: './visualization-list.component.html',
  styleUrls: ['./visualization-list.component.scss']
})
export class VisualizationListComponent implements OnInit {
  @ViewChildren(SortableDirective) headers: QueryList<SortableDirective>;
  visualizations: VisualizationType[] = [];
  fields: SortByType[] = [
    {
      fieldName: 'Name',
      field: 'name'
    },
    {
      fieldName: 'Last modification',
      field: 'modifiedDate'
    }
  ];
  loading = true;
  totalItems: any;
  page = 1;
  itemsPerPage = ITEMS_PER_PAGE;
  search: any;
  sortBy: string;
  selected: VisualizationType[] = [];
  allPageSelected: boolean;
  sort: SortEvent;
  private requestParams: any;

  constructor(private modalService: NgbModal,
              private visualizationService: VisualizationService,
              private spinner: NgxSpinnerService,
              private router: Router,
              private activatedRoute: ActivatedRoute) {
  }

  ngOnInit() {
    this.requestParams = {
      page: this.page - 1,
      size: this.itemsPerPage,
      sort: this.sortBy,
    };
    this.getVisualizationList();
    this.activatedRoute.queryParams.subscribe(params => {
      if (params[ActionInitParamsEnum.ON_INIT_ACTION]
        && params[ActionInitParamsEnum.ON_INIT_ACTION] === ActionInitParamsValueEnum.SHOW_CREATE_MODAL) {
        this.newVisualization();
      }
    });
  }

  editVisualization(vis: VisualizationType) {
    this.spinner.show('loadingSpinner');
    const queryParams = {};
    queryParams[VisualizationQueryParamsEnum.PATTERN_NAME] = vis.pattern.pattern;
    queryParams[VisualizationQueryParamsEnum.PATTERN_ID] = vis.pattern.id;
    queryParams[VisualizationQueryParamsEnum.CHART] = vis.chartType;
    queryParams[VisualizationQueryParamsEnum.MODE] = 'edit';
    queryParams[VisualizationQueryParamsEnum.VISUALIZATION_ID] = vis.id;
    const route = vis.chartType === ChartTypeEnum.TEXT_CHART ?
      '/creator/builder/text-builder' : '/creator/builder/chart-builder';
    this.router.navigate([route],
      {
        queryParams
      }).then(() => {
      this.spinner.hide('loadingSpinner');
    });
  }

  newVisualization() {
    this.modalService.open(VisualizationCreateComponent,
      {centered: true, size: 'lg'});
  }

  deleteVisualization(vis: VisualizationType) {
    const modal = this.modalService.open(VisualizationDeleteComponent, {centered: true});
    modal.componentInstance.visualization = vis;
    modal.componentInstance.visualizationDeleted.subscribe(() => {
      this.selected.splice(this.selected.indexOf(vis), 1);
      this.getVisualizationList();
    });
  }

  loadPage(page: any) {
    this.requestParams.page = page - 1;
    this.getVisualizationList();
  }

  chartIconResolver(chartType: string) {
    return UTM_CHART_ICONS[chartType];
  }

  getVisualizationList() {
    this.visualizationService.query(this.requestParams).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  importVisualizations() {
    const modalImport = this.modalService.open(VisualizationImportComponent, {centered: true});
    modalImport.componentInstance.visualizationImported.subscribe(() => {
      this.getVisualizationList();
    });
  }

  toggleAllSelection() {
    this.allPageSelected = !this.allPageSelected;
    if (this.allPageSelected) {
      this.visualizations.forEach(value => {
        if (this.selected.findIndex(sel => sel.id === value.id) === -1) {
          this.selected.push(value);
        }
      });
    } else {
      this.selected = [];
    }
  }

  isSelected(id: number) {
    return this.selected.findIndex(value => value.id === id) !== -1;
  }

  addToSelected(vis: VisualizationType) {
    const index = this.selected.findIndex(value => value.id === vis.id);
    if (index === -1) {
      this.selected.push(vis);
    } else {
      this.selected.splice(index, 1);
    }

  }

  exportVisualizations() {
    const dataStr = JSON.stringify(this.selected);
    const dataUri = 'data:application/json;charset=utf-8,' + encodeURIComponent(dataStr);
    const exportFileDefaultName = 'UTMVisualizations-' + moment(new Date()).format('YYYY-MM-DD') + '.json';
    const linkElement = document.createElement('a');
    linkElement.setAttribute('href', dataUri);
    linkElement.setAttribute('download', exportFileDefaultName);
    linkElement.click();
  }

  onFilterVisChange($event: any) {
    Object.keys($event).forEach(key => {
      if ($event[key] !== '' && $event[key] !== null) {
        this.requestParams[key] = $event[key];
      } else {
        this.requestParams[key] = undefined;
      }
    });
    this.getVisualizationList();
  }

  onSort($event: SortEvent) {
    this.sort = $event;
    this.requestParams.sort = $event.column + ',' + $event.direction;
    this.getVisualizationList();
  }

  bulkDelete() {
    const modal = this.modalService.open(VisualizationDeleteComponent, {centered: true});
    modal.componentInstance.selected = this.selected;
    modal.componentInstance.multiple = true;
    modal.componentInstance.visualizationDeleted.subscribe(() => {
      this.selected = [];
      this.getVisualizationList();
    });
  }

  private onSuccess(data, headers) {
    this.totalItems = headers.get('X-Total-Count');
    this.visualizations = data;
    this.loading = false;
  }

  private onError(error) {
    // this.alertService.error(error.error, error.message, null);
  }
}
