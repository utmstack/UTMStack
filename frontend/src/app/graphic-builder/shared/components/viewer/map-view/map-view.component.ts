import {AfterViewInit, Component, EventEmitter, Input, OnDestroy, OnInit, Output} from '@angular/core';
import {UUID} from 'angular2-uuid';
import * as L from 'leaflet';
import {Observable, Observer, of, Subject} from 'rxjs';
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
import {RefreshService, RefreshType} from "../../../../../shared/services/util/refresh.service";
import {catchError, filter, map, switchMap, takeUntil, tap} from "rxjs/operators";

@Component({
  selector: 'app-map-view',
  templateUrl: './map-view.component.html',
  styleUrls: ['./map-view.component.scss']
})
export class MapViewComponent implements OnInit, AfterViewInit, OnDestroy {
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
  data$: Observable<{ name: string, value: number[] }[]>;
  mapOption: LeafletMapType;
  map: any;
  contentLoaded: Observer<boolean>;
  sub = new Subject<boolean>();
  private chartEnumType: ChartTypeEnum = ChartTypeEnum.MARKER_CHART;
  error = false;
  defaultTime: ElasticFilterDefaultTime;
  markersLayer: any;
  refreshType: string;
  destroy$: Subject<void> = new Subject<void>();
  // tslint:disable-next-line:max-line-length
  res: { name: string, value: number[] }[] = [
    {
      "name": "57.135.189.63",
      "value": [
        26.2729,
        -80.26,
        28.0
      ]
    },
    {
      "name": "2601:58a:8984:9240:6463:35f3:3a5f:c61f",
      "value": [
        37.751,
        -97.822,
        22.0
      ]
    },
    {
      "name": "2601:58a:8984:9240:5c36:dc9f:f881:d87c",
      "value": [
        37.751,
        -97.822,
        9.0
      ]
    },
    {
      "name": "88.196.79.239",
      "value": [
        59.433,
        24.7323,
        7.0
      ]
    },
    {
      "name": "2601:58a:8984:9240:c08d:edc:4c08:ca06",
      "value": [
        37.751,
        -97.822,
        6.0
      ]
    },
    {
      "name": "2601:58a:8984:9240:a015:cc03:ee94:b65a",
      "value": [
        37.751,
        -97.822,
        4.0
      ]
    },
    {
      "name": "2601:58a:8984:9240:b509:9ea9:cb8c:ba47",
      "value": [
        37.751,
        -97.822,
        4.0
      ]
    },
    {
      "name": "2601:58a:8984:9240:1c73:b69d:5b2b:5b81",
      "value": [
        37.751,
        -97.822,
        3.0
      ]
    },
    {
      "name": "2601:58a:8984:9240:5863:8f23:93eb:4f60",
      "value": [
        37.751,
        -97.822,
        3.0
      ]
    },
    {
      "name": "2601:58a:8984:9240:6182:367b:f8f5:99cb",
      "value": [
        37.751,
        -97.822,
        3.0
      ]
    },
    {
      "name": "2601:58a:8984:9240:6cca:ff0:72c2:98e1",
      "value": [
        37.751,
        -97.822,
        3.0
      ]
    },
    {
      "name": "2601:58a:8984:9240:916a:63e6:9196:57ee",
      "value": [
        37.751,
        -97.822,
        3.0
      ]
    },
    {
      "name": "2601:58a:8984:9240:9587:4412:99a:4169",
      "value": [
        37.751,
        -97.822,
        3.0
      ]
    },
    {
      "name": "2601:58a:8984:9240:9ddb:ce24:5ea2:2f84",
      "value": [
        37.751,
        -97.822,
        3.0
      ]
    },
    {
      "name": "2601:58a:8984:9240:b13f:97f3:2728:950",
      "value": [
        37.751,
        -97.822,
        3.0
      ]
    },
    {
      "name": "2600:1700:234c:ac10:d0cb:fe60:ca36:8c59",
      "value": [
        25.6958,
        -80.3626,
        2.0
      ]
    },
    {
      "name": "2601:58a:8984:9240:217b:7bdc:ab48:e41a",
      "value": [
        37.751,
        -97.822,
        2.0
      ]
    },
    {
      "name": "2601:58a:8984:9240:9d3d:d2e1:d3f2:b458",
      "value": [
        37.751,
        -97.822,
        2.0
      ]
    },
    {
      "name": "107.202.154.128",
      "value": [
        25.7689,
        -80.1946,
        1.0
      ]
    },
    {
      "name": "2601:58a:8984:9240:a589:cf0c:9737:f4c6",
      "value": [
        37.751,
        -97.822,
        1.0
      ]
    },
    {
      "name": "2601:58a:8984:9240:ccbc:1190:7f84:c0b8",
      "value": [
        37.751,
        -97.822,
        1.0
      ]
    }
  ];

  constructor(private runVisualizationService: RunVisualizationService,
              private runVisualizationBehavior: RunVisualizationBehavior,
              private utmChartClickActionService: UtmChartClickActionService,
              private dashboardBehavior: DashboardBehavior,
              private toastService: UtmToastService,
              private refreshService: RefreshService) {

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
    this.refreshType = `${this.chartId}`;
    this.data$ = this.refreshService.refresh$
      .pipe(
        filter((refreshType) => refreshType === RefreshType.ALL ||
          refreshType === this.refreshType),
        switchMap((value, index) => this.runVisualization()));

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
    this.runVisualizationBehavior.$run
      .pipe(takeUntil(this.destroy$))
      .subscribe(id => {
      if (id && this.chartId === id) {
        this.refreshService.sendRefresh(this.refreshType);
        this.defaultTime = resolveDefaultVisualizationTime(this.visualization);
      }
    });
    this.dashboardBehavior.$filterDashboard
      .pipe(takeUntil(this.destroy$))
      .subscribe(dashboardFilter => {
      if (dashboardFilter && dashboardFilter.indexPattern === this.visualization.pattern.pattern) {
        mergeParams(dashboardFilter.filter, this.visualization.filterType).then(newFilters => {
          this.visualization.filterType = sanitizeFilters(newFilters);
          this.refreshService.sendRefresh(this.refreshType);
        });
      }
    });

    this.defaultTime = resolveDefaultVisualizationTime(this.visualization);
    if(!this.defaultTime){
      this.refreshService.sendRefresh(this.refreshType);
    }
  }

  runVisualization() {
    this.loadingOption = true;
    /*this.runVisualizationService.run(this.visualization).subscribe(resp => {
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
    });*/

    return this.runVisualizationService.run(this.visualization)
      .pipe(
        tap((data) => {
          this.loadingOption = false;
          this.runned.emit('runned');
          this.data = data;
          this.onChartChange();
          this.error = false;
        }),
        map( data => data.length > 0 ? data[0] : [] ),
        catchError(() => {
          this.loadingOption = false;
          this.error = true;
          this.runned.emit('runned');
          this.toastService.showError('Error',
            'Error occurred while running visualization');
          return of([]);
        })
      );
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
      this.refreshService.sendRefresh(this.refreshType);
    });
  }

  initMap() {
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

  ngOnDestroy(): void {
    this.refreshService.stopInterval();
    this.destroy$.next();
    this.destroy$.complete();
  }
}
