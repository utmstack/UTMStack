import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {VsSeverityResolverService} from '../../../../shared/services/scan/vs-severity-resolver.service';
import {SEVERITY_VALUE} from '../../../shared/const/task-status.const';
import {ScannerSeverityEnum} from '../../../shared/enums/scanner-severity.enum';
import {AssetModel} from '../../../shared/model/assets/asset.model';
import {AssetSoResolverService} from '../../../shared/providers/asset-so-resolver.service';

@Injectable({
  providedIn: 'root'
})
export class HostTopologyDef {
  severities = SEVERITY_VALUE;
  config: TopologyHost = {
    nodes: [],
    links: [],
  };

  constructor(private severityResolver: VsSeverityResolverService,
              private assetSoResolverService: AssetSoResolverService) {
  }

  public buildChartHostTopologyDef(assets: AssetModel[]) {
    let topology: any = {};
    this.processHostTopologyDef(assets).subscribe(topologyOption => {

      topology = {
        color: [
          '#90A4AE',
          '#00838F',
          '#ff9800',
          '#C62828',
          '#343a40'],
        toolbox: {
          show: true,
          feature: {
            restore: {
              title: 'Reset',
              show: true
            },
          }
        },
        tooltip: {
          show: true,
          // formatter: '{b}',
          formatter: (params) => {
            // if (params.data.name === '>') {
            const assetIndex = assets.findIndex(value => params.data.name === value.name);
            if (assetIndex !== -1) {
              const host = assets[assetIndex];
              return this.buildTooltip(host);
            } else {
              return params.data.name;
            }
          },
          trigger: 'item',
          backgroundColor: 'rgba(0,75,139,0.85)',
          padding: 10,
          textStyle: {
            fontSize: 13,
            fontFamily: 'Roboto, sans-serif'
          }
        },
        grid: {
          top: '35px',
          left: 0,
          right: 0,
          bottom: 0
        },
        series: [
          {
            type: 'graph',
            layout: 'force',
            force: {
              repulsion: 200,
              layoutAnimation: true
            },
            circular: {
              rotateLabel: false
            },
            draggable: true,
            name: 'Host',
            ribbonType: true,
            itemStyle: {
              normal: {
                label: {
                  position: 'top',
                  distance: 5,
                  show: true,
                  textStyle: {
                    margin: '15px',
                    fontSize: '9px',
                    color: '#333'
                  }
                },
                nodeStyle: {
                  brushType: 'both',
                },
                linkStyle: {
                  type: 'curve'
                }
              },
              emphasis: {
                label: {
                  show: false
                  // textStyle: null      // 默认使用全局文本样式，详见TEXTSTYLE
                },
                nodeStyle: {
                  // r: 30
                },
                linkStyle: {
                  type: 'line',
                  color: '#333'
                }
              }
            },
            useWorker: true,
            minRadius: 15,
            categories: [
              {
                name: ScannerSeverityEnum.UNKNOWN,
                itemStyle: {
                  color: '#343a40',
                },
              },
              {
                name: ScannerSeverityEnum.LOG,
                itemStyle: {
                  color: '#90A4AE',
                },
              },
              {
                name: ScannerSeverityEnum.LOG,
                itemStyle: {
                  color: '#00838F',
                },
              },
              {
                name: ScannerSeverityEnum.MEDIUM,
                itemStyle: {
                  color: '#ff9800',
                },
              },
              {
                name: ScannerSeverityEnum.HIGH,
                itemStyle: {
                  color: '#C62828',
                },
              }
            ],
            maxRadius: 25,
            gravity: 1.1,
            scaling: 0.5,
            roam: 'scale',
            edgeSymbol: 'arrow',
            edgeSymbolSize: 5,
            // focusNodeAdjacency: true,
            symbol: 'path://M67.998,0H21.473c-1.968,0-3.579,1.61-3.579,3.579v82.314c0,1.968,1.61,3.579,3.579,3.579h46.525' +
              'c1.968,0,3.579-1.61,3.579-3.579V3.579C71.577,1.61,69.967,0,67.998,0z M44.736,65.811c-2.963,0-5.368-2.409-5.368-5.368' +
              'c0-2.963,2.405-5.368,5.368-5.368c2.963,0,5.368,2.405,5.368,5.368C50.104,63.403,47.699,65.811,44.736,65.811z M64.419,39.704' +
              'H25.052v-1.789h39.367V39.704z M64.419,28.967H25.052v-1.789h39.367V28.967z M64.419,17.336H25.052V6.599h39.367V17.336z',
            symbolSize: '15',
            symbolKeepAspect: true,
            nodeScaleRatio: 1,
            lineStyle: {
              color: '#777'
            },
            nodes: topologyOption.nodes,
            links: topologyOption.links
          }]
      };
    });
    return topology;
  }

  buildTooltip(asset: AssetModel): string {
    return '<div class="so d-flex align-items-center justify-content-start flex-column"' +
      '<span class="pl-1">' +
      'Host name: ' +
      asset.name +
      '</span>' +
      '<span class="pl-1 pt-2">' +
      'OS: ' +
      this.assetSoResolverService.resolveAssetSoName(asset) +
      '</span>' +
      '<span class="pl-1 pt-2">' +
      'Severity: ' +
      this.severityResolver.resolveSeverityLabel(asset.host.severity.value) +
      '</span>';
  }

  private processHostTopologyDef(assets: AssetModel[]): Observable<TopologyHost> {
    return new Observable<TopologyHost>((data) => {
      if (assets !== null) {
        for (const asset of assets) {
          if (asset.host.detail) {
            const index = asset.host.detail.findIndex(value => value.name === 'traceroute');
            if (index !== -1) {
              const traceroute: string[] = asset.host.detail[index].value.split(',');
              // tslint:disable-next-line:prefer-for-of
              for (let i = 0; i < traceroute.length; i++) {
                const indexNode = this.config.nodes.findIndex(value => traceroute[i] === value.name);
                const indexAsset = assets.findIndex(value => value.name === traceroute[i]);
                const assetName = indexAsset !== -1 ? assets[indexAsset].host.severity.value : null;
                const severityIndex = this.severities.findIndex(value =>
                  value === this.severityResolver.resolveSeverityLabel(assetName));
                if (indexNode === -1) {
                  this.config.nodes.push({
                    category: severityIndex,
                    name: traceroute[i],
                    value: 1
                  });
                } else {
                  if (this.config.nodes[indexNode].category < severityIndex) {
                    this.config.nodes[indexNode].category = severityIndex;
                  }
                }
                if (traceroute[i + 1]) {
                  this.config.links.push({
                    source: traceroute[i],
                    target: traceroute[i + 1],
                    weight: this.severities.findIndex(value =>
                      value === this.severityResolver.resolveSeverityLabel(asset.host.severity.value))
                  });
                }
              }
            }
          }
        }
        data.next(this.config);
      } else {
        data.next(null);
      }
    });
  }


}


export class TopologyHost {
  nodes: {
    category?: number, name?: string, value?: number, itemStyle?: {
      color?: string
    }
  }[];
  links: { source?: string, target?: string, weight?: number }[];
}
