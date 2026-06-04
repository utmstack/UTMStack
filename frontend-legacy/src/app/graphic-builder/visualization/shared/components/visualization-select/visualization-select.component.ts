import {HttpResponse} from '@angular/common/http';
import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {VisualizationType} from '../../../../../shared/chart/types/visualization.type';
import {UTM_CHART_ICONS} from '../../../../../shared/constants/icons-chart.const';
import {RouteCallbackEnum} from '../../../../../shared/enums/route-callback.enum';
import {IndexPatternService} from '../../../../../shared/services/elasticsearch/index-pattern.service';
import {UtmIndexPattern} from '../../../../../shared/types/index-pattern/utm-index-pattern';
import {SortByType} from '../../../../../shared/types/sort-by.type';
import {CHART_TYPES} from '../../../../shared/const/chart-type.const';
import {VisualizationCreateComponent} from '../../../visualization-create/visualization-create.component';
import {VisualizationService} from '../../services/visualization.service';

@Component({
  selector: 'app-visualization-select',
  templateUrl: './visualization-select.component.html',
  styleUrls: ['./visualization-select.component.scss']
})
export class VisualizationSelectComponent implements OnInit {
  @Input() emitOnSelect: boolean;
  @Input() dataNature: number;
  @Input() callback: RouteCallbackEnum;
  @Output() visSelected = new EventEmitter<VisualizationType[]>();
  @Output() visClose = new EventEmitter<string>();
  @Output() createNewVisualization = new EventEmitter<boolean>();
  charts = CHART_TYPES;
  visualizations: VisualizationType[] = [];
  fields: SortByType[] = [
    {
      fieldName: 'Name',
      field: 'name'
    },
    {
      fieldName: 'Last modification',
      field: 'modificationTime'
    }
  ];
  loading = true;
  totalItems: any;
  page = 1;
  itemsPerPage = 15;
  search: any;
  sortBy: string;
  visualizationSearch: string;
  visualizationSelected: VisualizationType[] = [];
  searching = false;
  chartType: string;
  patterns: UtmIndexPattern[];
  constructor(private modalService: NgbModal,
              private indexPatternService: IndexPatternService,
              private visualizationService: VisualizationService) {
  }

  ngOnInit() {
    this.getIndexPatterns();
    this.getVisualizationList();
  }


  onSortBy($event) {
  }

  chartIconResolver(chartType: string) {
    return UTM_CHART_ICONS[chartType];
  }

  loadPage(page: any) {
    this.page = page;
    this.getVisualizationList();
  }

  searchVisualization($event: any) {
    this.searching = true;
    this.search = $event;
    this.getVisualizationList();
  }

  addToSelected(vis) {
    const index = this.visualizationSelected.findIndex(value => value.id === vis.id);
    if (index === -1) {
      this.visualizationSelected.push(vis);
    } else {
      this.visualizationSelected.splice(index, 1);
    }
    if (this.emitOnSelect) {
      this.visSelected.emit(this.visualizationSelected);
    }
  }

  isSelected(vis): boolean {
    return this.visualizationSelected.findIndex(value => value.name === vis.name) > -1;
  }

  cancelVisualization() {
    this.visClose.emit('cancel');
  }

  addingVisualization() {
    this.visSelected.emit(this.visualizationSelected);
  }

  getVisualizationList() {
    const req = {
      page: this.page - 1,
      size: this.itemsPerPage,
      'name.contains': this.search,
      sort: this.sortBy,
      'idPattern.equals': this.dataNature,
      'chartType.equals': this.chartType
    };
    this.visualizationService.query(req).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  newVisualization() {
    this.createNewVisualization.emit(true);
    const modal = this.modalService.open(VisualizationCreateComponent,
      {centered: true, size: 'lg'});
    modal.componentInstance.callback = this.callback;
  }

  private onSuccess(data, headers) {
    this.totalItems = headers.get('X-Total-Count');
    this.visualizations = data;
    this.loading = false;
    this.searching = false;
  }

  private onError(error) {
    // this.alertService.error(error.error, error.message, null);
  }

  filterByIndexPattern($event: any) {
    if ($event) {
      this.dataNature = $event.id;
    } else {
      this.dataNature = null;
    }

    this.getVisualizationList();
  }

  getIndexPatterns() {
    const req = {
      page: 0,
      size: 1000,
    };
    this.indexPatternService.query(req).subscribe(
      (res: HttpResponse<any>) => {
        this.patterns = res.body;
      },
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }
}
