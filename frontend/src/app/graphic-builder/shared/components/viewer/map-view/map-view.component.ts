import {AfterViewInit, Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {UUID} from 'angular2-uuid';
import * as L from 'leaflet';
import {Observable, Observer, Subject} from 'rxjs';
import {UtmToastService} from '../../../../../shared/alert/utm-toast.service';
import {DashboardBehavior} from '../../../../../shared/behaviors/dashboard.behavior';
import {EchartClickAction} from '../../../../../shared/chart/types/action/echart-click-action';
import {UtmScatterMapOptionType} from '../../../../../shared/chart/types/charts/scatter/utm-scatter-map-option.type';
import {LeafletMapType} from '../../../../../shared/chart/types/map/leaflet/leaflet-map.type';
import {VisualizationType} from '../../../../../shared/chart/types/visualization.type';
import {ElasticFilterDefaultTime} from '../../../../../shared/components/utm/filters/elastic-filter-time/elastic-filter-time.component';
import {ChartTypeEnum} from '../../../../../shared/enums/chart-type.enum';
import {TimeFilterType} from '../../../../../shared/types/time-filter.type';
import {mergeParams, sanitizeFilters} from '../../../../../shared/util/elastic-filter.util';
import {getBucketLabel} from '../../../../chart-builder/chart-property-builder/shared/functions/visualization-util';
import {RunVisualizationBehavior} from '../../../behavior/run-visualization.behavior';
import {RunVisualizationService} from '../../../services/run-visualization.service';
import {UtmChartClickActionService} from '../../../services/utm-chart-click-action.service';
import {rebuildVisualizationFilterTime} from '../../../util/chart-filter/chart-filter.util';
import {resolveDefaultVisualizationTime} from '../../../util/visualization/visualization-render.util';

@Component({
  selector: 'app-map-view',
  templateUrl: './map-view.component.html',
  styleUrls: ['./map-view.component.scss']
})
export class MapViewComponent implements OnInit, AfterViewInit {
  mapId = UUID.UUID();
  @Input() building: boolean;
  @Input() chartId: number;
  @Input() chart: ChartTypeEnum;
  @Input() visualization: VisualizationType;
  @Input() width: string;
  @Input() height: string;
  @Input() showTime: boolean;
  @Input() timeByDefault: any;
  @Output() runned = new EventEmitter<string>();
  loadingOption = true;
  data: { name: string, value: number[] }[] = [];
  mapOption: LeafletMapType;
  map: any;
  contentLoaded: Observer<boolean>;
  sub = new Subject<boolean>();
  private chartEnumType: ChartTypeEnum = ChartTypeEnum.MARKER_CHART;
  error = false;
  defaultTime: ElasticFilterDefaultTime;
  markersLayer: any;

  constructor(private runVisualizationService: RunVisualizationService,
              private runVisualizationBehavior: RunVisualizationBehavior,
              private utmChartClickActionService: UtmChartClickActionService,
              private dashboardBehavior: DashboardBehavior,
              private toastService: UtmToastService) {

  }

  mapLoaded(): Observable<boolean> {
    return this.sub.asObservable();
  }

  ngAfterViewInit(): void {
    // const osmLink = '<a href="http://openstreetmap.org">OpenStreetMap</a>';
    // let osmUrl: string;
    // let osmAttrib: string;
    // if (this.mapOption && this.mapOption.tiles.length === 0) {
    //   osmUrl = 'http://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png';
    //   osmAttrib = '&copy; ' + osmLink + ' Contributors';
    // } else {
    //   osmUrl = this.mapOption.tiles[0].urlTemplate;
    //   osmAttrib = this.mapOption.tiles[0].options.attribution;
    // }
    // const osmMap = L.tileLayer(osmUrl, {attribution: osmAttrib});
    this.map = new L.map(this.mapId, {
      // layers: [osmMap], // only add one!
      minZoom: 1
    });
    this.markersLayer = new L.LayerGroup();
    this.sub.next(true);

  }

  ngOnInit() {
    this.mapLoaded().subscribe(value => {
      if (value) {
        const baseLayers = {};
        if (typeof this.visualization.chartConfig === 'string') {
          JSON.parse(this.visualization.chartConfig);
        }
        const config: UtmScatterMapOptionType = JSON.parse(this.visualization.chartConfig);
        const option: LeafletMapType = config.leaflet;
        for (const tile of option.tiles) {
          baseLayers[tile.label] = L.tileLayer(tile.urlTemplate, {attribution: tile.options.attribution});
        }
        const layers = L.control.layers(baseLayers);
        layers.addTo(this.map);
        for (const tile of option.tiles) {
          L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">' +
              'OpenStreetMap</a> contributors'
          }).addTo(this.map);
        }
        setTimeout(() => {
          this.map.invalidateSize();
        }, 1);
      }
    });
    this.runVisualizationBehavior.$run.subscribe(id => {
      if (id && this.chartId === id) {
        this.runVisualization();
        this.defaultTime = resolveDefaultVisualizationTime(this.visualization);
      }
    });
    this.dashboardBehavior.$filterDashboard.subscribe(dashboardFilter => {
      if (dashboardFilter && dashboardFilter.indexPattern === this.visualization.pattern.pattern) {
        mergeParams(dashboardFilter.filter, this.visualization.filterType).then(newFilters => {
          this.visualization.filterType = sanitizeFilters(newFilters);
          this.runVisualization();
        });
      }
    });
    this.runVisualization();
    this.defaultTime = resolveDefaultVisualizationTime(this.visualization);
  }

  runVisualization() {
    this.loadingOption = true;
    this.runVisualizationService.run(this.visualization).subscribe(resp => {
      this.loadingOption = false;
      this.runned.emit('runned');
      this.data = resp;
      this.onChartChange();
      this.error = false;
    }, error => {
      this.loadingOption = false;
      this.error = true;
      this.runned.emit('runned');
      this.toastService.showError('Error',
        'Error occurred while running visualization');
    });
  }

  /**
   * Build echart object
   */
  onChartChange() {
    // set center
    this.mapOption = this.visualization.chartConfig.leaflet;
    const centerLat = (this.data && this.data.length > 0) ? this.data[0].value[0] : this.mapOption.center[0];
    const centerLng = (this.data && this.data.length > 0) ? this.data[0].value[1] : this.mapOption.center[1];

    this.map.setView([centerLat, centerLng], this.mapOption.zoom);
    this.map.panTo([centerLat, centerLng]);



    this.markersLayer.clearLayers();

    if (this.data && this.data.length > 0) {
      this.data.sort((a, b) => a.value[2] < b.value[2] ? 1 : -1);
    }

    if (this.data && this.data.length > 0) {
      for (const dat of this.data) {
        const tooltip = '<div style="' +
          'white-space: nowrap;' +
          'padding-top:10px' +
          'font: 13px / 20px Poppins, sans-serif;' +
          'pointer-events: none;">' +
          getBucketLabel(0, this.visualization) +
          '<br>' +
          '<span style="display:inline-block;' +
          'margin-right:5px;' +
          'border-radius:10px;' +
          'width:10px;' +
          'height:10px;' +
          'background-color:#c05050;">' +
          '</span><strong>' + dat.name + ':</strong> ' + dat.value[2] + '</div>';
        const myIcon = new L.Icon({
          iconUrl: '/assets/img/red_marker.png',
          iconSize: [15, 25],
          iconAnchor: [4, 25],
          popupAnchor: [4, -30],
        });
        const marker = L.marker([dat.value[0], dat.value[1]], {icon: myIcon})
          .bindPopup(tooltip, {className: 'class-popup'});

        this.markersLayer.addLayer(marker);

        marker.on('mouseover', (ev) => {
          marker.openPopup();
        });
        marker.on('mouseout', (ev) => {
          marker.closePopup();
        });
        marker.on('click', (ev) => {
          const echartClickAction: EchartClickAction = {
            data: {name: dat.name}
          };
          if (!this.building) {
            this.utmChartClickActionService.onClickNavigate(this.visualization, echartClickAction);
          }
        });
      }
    }

    setTimeout(() => {
      this.map.addLayer(this.markersLayer);
      this.map.invalidateSize();
    }, 1);
  }

  eConsole(param) {
    if (typeof param.seriesIndex !== 'undefined') {
    }
  }

  convertData(dat, geoCoordMap) {
    const res = [];
    // tslint:disable-next-line:prefer-for-of
    for (let i = 0; i < dat.length; i++) {
      const geoCoord = geoCoordMap[dat[i].name];
      if (geoCoord) {
        res.push({
          name: dat[i].name,
          value: geoCoord.concat(dat[i].value)
        });
      }
    }
    return res;
  }

  onTimeFilterChange($event: TimeFilterType) {
    rebuildVisualizationFilterTime($event, this.visualization.filterType).then(filters => {
      this.visualization.filterType = filters;
      this.runVisualization();
    });
  }
}
