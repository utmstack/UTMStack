import {MapTiles} from '../../../types/charts/chart-properties/map-tiles/map-tiles';
import {SeriesScatter} from '../../../types/charts/chart-properties/series/scatter/series-scatter';
import {UtmScatterMapOptionType} from '../../../types/charts/scatter/utm-scatter-map-option.type';
import {LeafletTilesType} from '../../../types/map/leaflet/leaflet-tiles.type';
import {ScatterMapCoordinateResponseType} from '../../../types/response/scatter-map-response.type';
import {VisualizationType} from '../../../types/visualization.type';
import {ChartBuildInterface} from '../chart-build.interface';
import {ChartOption} from '../chart-option';

export class ScatterMap implements ChartBuildInterface {

  constructor() {
  }


  reformatTiles(tiles: MapTiles[]): Promise<LeafletTilesType[]> {
    return new Promise<LeafletTilesType[]>(resolve => {
      const tilesScatter: LeafletTilesType[] = [];
      for (const tile of tiles) {
        tilesScatter.push({
          label: tile.label,
          urlTemplate: tile.urlTemplate,
          options: {
            attribution: tile.attribution ? tile.attribution : 'Utm Vault maps'
          }
        });
      }
      resolve(tilesScatter);
    });
  }

  buildChart(data?: ScatterMapCoordinateResponseType[], visualization?: VisualizationType): ChartOption {
    if (data && visualization.chartConfig) {
      const scatterOptions: UtmScatterMapOptionType = visualization.chartConfig;
      this.resolveSeries(data, visualization).then(series => {
        scatterOptions.series = series;
      });
      scatterOptions.leaflet.tiles =
        [
          {
            label: 'Open Street Map',
            urlTemplate: 'https://{s}.tile.openstreetmap.fr/hot/{z}/{x}/{y}.png',
            options: {
              attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
            }
          }
        ];
      scatterOptions.leaflet.center = [data[0].value[0], data[0].value[1]];
      return scatterOptions;
    } else {
      return null;
    }
  }

  resolveSeries(data: ScatterMapCoordinateResponseType[], visualization?: VisualizationType): Promise<SeriesScatter[]> {
    return new Promise<SeriesScatter[]>(resolve => {
      const options: UtmScatterMapOptionType = visualization.chartConfig;
      const normalSeries: SeriesScatter = {
        type: 'scatter',
        coordinateSystem: 'leaflet',
        name: '',
        data: [
          {
            name: 'Ip 21',
            value: [20.8795129, -76.2594977, 50, 50]
          }
        ],
        symbolSize: (val) => {
          return val[2] / 100;
        },
        rippleEffect: {
          brushType: 'fill'
        },
        showEffectOn: 'render',
        hoverAnimation: true,
        itemStyle: {
          normal: {
            color: '#f5994e'
          }
        }
      };
      const highlightSeries: SeriesScatter = {
        name: 'Top ' + options.top,
        type: 'effectScatter',
        coordinateSystem: 'leaflet',
        data: (data.sort((a, b) => {
          return a.value > b.value ? 1 : -1;
        }).slice(0, options.top)),
        symbolSize: (val) => {
          return val[2] / 100;
        },
        showEffectOn: 'render',
        rippleEffect: {
          brushType: 'stroke'
        },
        hoverAnimation: true,
        zlevel: 1,
        itemStyle: {
          normal: {
            color: '#c05050',
            shadowBlur: 10,
            shadowColor: '#333'
          }
        },
      };
      if (data.length > options.top) {
        resolve([normalSeries, highlightSeries]);
      } else {
        resolve([normalSeries]);
      }
    });
  }
}
