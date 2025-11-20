import {Component, Input, OnChanges, OnInit} from '@angular/core';
import {AdversaryAlerts} from '../models';

@Component({
  selector: 'app-adversary-alerts-graph',
  templateUrl: './adversary-alerts-graph.component.html',
  styleUrls: ['./adversary-alerts-graph.component.scss']
})
export class AdversaryAlertsGraphComponent implements OnChanges {
  @Input() data!: AdversaryAlerts[];

  option: any;

  ngOnChanges(): void {
    if (this.data) {
      this.option = this.buildOption(this.data);
      console.log(this.option);
    }
  }

  private buildOption(adversaryAlerts: AdversaryAlerts[]): any {
    const nodes: any[] = [];
    const links: any[] = [];
    const nodeSet = new Set<string>();
    const colorMap = new Map<string, string>();
    const adversaryColors = ['#5470C6', '#91CC75', '#EE6666', '#FAC858', '#73C0DE', '#3BA272', '#FC8452'];
    let colorIndex = 0;

    const nodeKey = (id: string, name: string) => `${id}::${name}`;
    const truncate = (text: string, max = 30) => text.length > max ? text.slice(0, max) + '…' : text;

    adversaryAlerts.forEach(group => {
      const advId = group.adversary.host ? group.adversary.host
        : group.adversary.user ? group.adversary.user : 'adv';
      const advName = truncate(advId);
      const advKey = nodeKey(advId, advName);
      const advColor = adversaryColors[colorIndex++ % adversaryColors.length];
      colorMap.set(advKey, advColor);

      if (!nodeSet.has(advKey)) {
        nodes.push({
          name: advKey,
          label: { formatter: advName },
          itemStyle: { color: advColor }
        });
        nodeSet.add(advKey);
      }

      group.alerts.forEach(alertWithChildren => {
        const alertKey = nodeKey(alertWithChildren.alert.id, alertWithChildren.alert.name);
        const alertLabel = truncate(alertWithChildren.alert.name);
        if (!nodeSet.has(alertKey)) {
          nodes.push({
            name: alertKey,
            label: { formatter: alertLabel },
            itemStyle: { color: advColor }
          });
          nodeSet.add(alertKey);
        }

        links.push({
          source: advKey,
          target: alertKey,
          value: alertWithChildren.children.length || 1,
          lineStyle: { color: advColor }
        });

        alertWithChildren.children.forEach(child => {
          const childKey = nodeKey(child.id, child.name);
          const childLabel = truncate(child.name);
          if (!nodeSet.has(childKey)) {
            nodes.push({
              name: childKey,
              label: { formatter: childLabel },
              itemStyle: { color: advColor }
            });
            nodeSet.add(childKey);
          }

          links.push({
            source: alertKey,
            target: childKey,
            value: 1,
            lineStyle: { color: advColor }
          });
        });
      });
    });

    return {
      title: { text: 'Adversaries → Alerts → Echoes', left: 'center' },
      tooltip: {
        trigger: 'item',
        formatter: (params: any) => params.name.split('::')[1] ? params.name.split('::')[1] : params.name
      },
      series: [{
        type: 'sankey',
        data: nodes,
        links,
        emphasis: { focus: 'adjacency' },
        nodeWidth: 20,
        nodeGap: 12,
        left: 80,
        right: 120,
        top: 60,
        bottom: 30
      }]
    };
  }

}
