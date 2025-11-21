import {Component, EventEmitter, Input, OnChanges, OnInit, Output} from '@angular/core';
import {ECharts} from 'echarts';
import {AdversaryAlerts} from '../models';

@Component({
  selector: 'app-adversary-alerts-graph',
  templateUrl: './adversary-alerts-graph.component.html',
  styleUrls: ['./adversary-alerts-graph.component.scss']
})
export class AdversaryAlertsGraphComponent implements OnChanges {
  @Input() data!: AdversaryAlerts[];
  chartHeight: number;
  baseHeight = 600;
  nodeGap = 6;
  option: any;

  ngOnChanges(): void {
    if (this.data) {
      this.option = this.buildOption(this.data);
    }
  }

  onChartInit(chart: ECharts) {
    // Paso 1: capturar mouseover en nodos
    chart.on('mouseover', (params) => {
      if (params.dataType === 'node') {
        chart.getDom().style.cursor = 'pointer';
      }
    });

    chart.on('mouseout', (params) => {
      if (params.dataType === 'node') {
        chart.getDom().style.cursor = 'default';
      }
    });

    // Extra: capturar click en nodo
    chart.on('click', (params) => {
      if (params.dataType === 'node') {
        console.log('Nodo clicado:', params.data);
        // aquí puedes abrir un modal o navegar a detalle
      }
    });
  }

  private buildOption(adversaryAlerts: AdversaryAlerts[]): any {
    const nodes: any[] = [];
    const links: any[] = [];
    const nodeSet = new Set<string>();
    const colorMap = new Map<string, string>();
    const severityColors = {
      high: '#D62728',    // red
      medium: '#FF7F0E',  // orange
      low: '#2CA02C',     // green
      default: '#1F77B4'  // blue
    };
    const adversaryColors = [
      '#1F77B4', '#FF7F0E', '#2CA02C', '#D62728',
      '#9467BD', '#8C564B', '#E377C2', '#7F7F7F',
      '#BCBD22', '#17BECF'
    ];
    let colorIndex = 0;

    const nodeKey = (id: string, name: string) => `${id}::${name}`;
    const truncate = (text: string, max = 30) => text.length > max ? text.slice(0, max) + '…' : text;
    const getNodeSize = (count: number, base = 10, max = 30) => Math.min(base + count * 2, max);
    const gradientColor = (from: string, to: string) => ({
      type: 'linear',
      x: 0, y: 0, x2: 1, y2: 0,
      colorStops: [
        { offset: 0, color: from },
        { offset: 1, color: to }
      ]
    });

    adversaryAlerts.forEach(group => {
      const advId = group.adversary.host || group.adversary.user || 'adv';
      const advName = truncate(advId);
      const advKey = nodeKey(advId, advName);
      const advColor = adversaryColors[colorIndex++ % adversaryColors.length];
      colorMap.set(advKey, advColor);

      if (!nodeSet.has(advKey)) {
        nodes.push({
          name: advKey,
          label: { formatter: advName },
          itemStyle: {
            color: advColor,
            borderColor: '#fff',
            borderWidth: 1,
            opacity: 0.9
          },
          depth: 0
        });
        nodeSet.add(advKey);
      }

      group.alerts.forEach(alertWithChildren => {
        const alert = alertWithChildren.alert;
        const alertKey = nodeKey(alert.id, alert.name);
        const alertLabel = truncate(alert.name);
        const childCount = alertWithChildren.children.length;
        const severityColor = severityColors[alert.severityLabel.toLowerCase()] || severityColors.default;

        if (!nodeSet.has(alertKey)) {
          nodes.push({
            name: alertKey,
            label: { formatter: alertLabel },
            itemStyle: {
              color: severityColor,
              borderColor: '#fff',
              borderWidth: 1,
              opacity: 0.9
            },
            depth: 1,
            symbolSize: getNodeSize(childCount),
            meta: {
              severity: alert.severityLabel,
              timestamp: alert.timestamp,
              dataSource: alert.dataSource
            }
          });
          nodeSet.add(alertKey);
        }

        links.push({
          source: advKey,
          target: alertKey,
          value: childCount || 1,
          lineStyle: {
            color: gradientColor(advColor, severityColor),
            opacity: 0.6,
            curveness: 0.3
          }
        });

        alertWithChildren.children.forEach(child => {
          const childKey = nodeKey(child.id, child.name);
          const childLabel = truncate(child.name);
          const childSeverityColor = severityColors[child.severityLabel.toLowerCase()] || severityColors.default;

          if (!nodeSet.has(childKey)) {
            nodes.push({
              name: childKey,
              label: { formatter: childLabel },
              itemStyle: {
                color: childSeverityColor,
                borderColor: '#eee',
                borderWidth: 1,
                opacity: 0.8
              },
              depth: 2,
              symbolSize: getNodeSize(1),
              meta: {
                severity: child.severityLabel,
                timestamp: child.timestamp,
                dataSource: child.dataSource
              }
            });
            nodeSet.add(childKey);
          }

          links.push({
            source: alertKey,
            target: childKey,
            value: 1,
            lineStyle: {
              color: gradientColor(severityColor, childSeverityColor),
              opacity: 0.5,
              curveness: 0.2
            }
          });
        });
      });
    });

    const totalNodes = nodes.length;
    this.chartHeight = this.baseHeight + totalNodes * this.nodeGap;

    return {
      tooltip: {
        trigger: 'item',
        formatter: (params: any) => {
          const name = params.name.split('::')[1] || params.name;
          const meta = params.data.meta;
          if (!meta) { return name; }

          return `
          <strong>${name}</strong><br/>
          ${meta.eventCode ? `Event Code: ${meta.eventCode}<br/>` : ''}
          ${meta.severity ? `Severity: ${meta.severity}<br/>` : ''}
          ${meta.timestamp ? `Timestamp: ${new Date(meta.timestamp).toLocaleString()}<br/>` : ''}
        `;
        }
      },
      series: [{
        type: 'sankey',
        orient: 'horizontal',
        data: nodes,
        links,
        emphasis: {
          focus: 'adjacency',
          label: {
            fontSize: 14,
            fontWeight: 'bold',
            color: '#000'
          },
          itemStyle: {
            opacity: 1,
            borderWidth: 2,
            borderColor: '#000'
          }
        },
        nodeWidth: 15,
        nodeGap: this.nodeGap,
        nodeAlign: 'left',
        layoutIterations: 32,
        left: 20,
        right: 180,
        top: 20,
        bottom: 40,
        label: {
          position: 'right',
          fontSize: 10,
          formatter: (params: any) => params.name.split('::')[1] || params.name
        }
      }]
    };
  }
}
