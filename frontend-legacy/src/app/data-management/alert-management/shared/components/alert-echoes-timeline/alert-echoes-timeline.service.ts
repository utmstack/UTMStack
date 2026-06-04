import { Injectable } from '@angular/core';
import {UtmAlertType} from '../../../../../shared/types/alert/utm-alert.type';
import {TimelineItem} from '../../../../../shared/types/utm-timeline-item';

export interface TimelineGroup {
  startTimestamp: number;
  items: TimelineItem[];
  yOffset?: number;
}

const cardWidth = 240;
const cardHeight = 62;
const baseOffset = 80;
const spacing = 10;


@Injectable()
export class AlertEchoesTimelineService {

  renderItem(params: any, api: any) {
    const ts = api.value(0);
    const coord = api.coord([ts, 0]) as number[];
    const chartWidth = params.coordSys.width;
    api.getHeight();

    // Horizontal position (centered, respecting canvas borders)
    let x = coord[0] - cardWidth / 2;
    x = Math.max(0, Math.min(x, chartWidth - cardWidth));

    // Vertical stacking offset
    const level = api.value(7) || 0; // stack level
    const levelOffset = level * (cardHeight + spacing);
    let yCard = coord[1] - baseOffset - levelOffset - cardHeight;

    // Ensure card stays within canvas
    if (yCard < 0) { yCard = 0; }

    // Generic type for children to satisfy TS
    const children: {
      type: string;
      shape?: Record<string, number | string>;
      style?: Record<string, number | string>;
    }[] = [];

    // Card background
    children.push({
      type: 'rect',
      shape: { x, y: yCard, width: cardWidth, height: cardHeight, r: 10 },
      style: {
        fill: '#ffffff',
        stroke: '#0277bd',
        lineWidth: 1,
        shadowBlur: 8,
        shadowColor: 'rgba(0,0,0,0.2)',
        cursor: 'pointer'
      }
    });

    // Icon
    children.push({
      type: 'image',
      style: {
        image: api.value(4),
        x: x + 5,
        y: yCard + 5,
        width: cardHeight - 10,
        height: cardHeight - 10
      }
    });

    // Title
    children.push({
      type: 'text',
      style: {
        x: x + (cardHeight - 2.5) + 15,
        y: yCard + 5,
        text: this.truncateText(api.value(2) || '', 150),
        textAlign: 'left',
        fill: '#000',
        fontSize: 14,
        fontWeight: 600,
        width: cardWidth - (cardHeight - 2.5) - 25,
        overflow: 'break',
        ellipsis: '...'
      }
    });

    // Subtitle / date
    children.push({
      type: 'text',
      style: {
        x: x + (cardHeight - 2.5) + 15,
        y: yCard + 25,
        text: api.value(3),
        textAlign: 'left',
        fill: '#666',
        fontSize: 12
      }
    });

    // Total echoes
    if (api.value(5) > 1) {
      children.push({
        type: 'text',
        style: {
          x: x + (cardHeight - 2.5) + 15,
          y: yCard + cardHeight - 18,
          text: `Total: ${api.value(5)} echoes`,
          textAlign: 'left',
          fill: '#444',
          fontSize: 12,
          fontWeight: 'bold'
        }
      });
    }

    // Line connecting to timeline
    children.push({
      type: 'line',
      shape: {
        x1: coord[0],
        y1: yCard + cardHeight,
        x2: coord[0],
        y2: coord[1]
      },
      style: {
        stroke: '#0277bd',
        lineWidth: 1.5
      }
    });

    return { type: 'group', children };
  }


  private groupByInterval(items: TimelineItem[], intervalMs: number): TimelineGroup[] {
    const sorted = [...items].sort(
      (a, b) => new Date(a.startDate).getTime() - new Date(b.startDate).getTime()
    );
    const groups: TimelineGroup[] = [];
    let currentGroup: TimelineGroup = null;

    sorted.forEach(item => {
      const ts = new Date(item.startDate).getTime();
      if (!currentGroup || ts > currentGroup.startTimestamp + intervalMs) {
        currentGroup = { startTimestamp: ts, items: [] };
        groups.push(currentGroup);
      }
      currentGroup.items.push(item);

      if (ts > currentGroup.startTimestamp) {
        currentGroup.startTimestamp = ts;
      }
    });

    return groups;
  }

  truncateText(text: string, maxWidth: number) {
    const avgCharWidth = 7;
    const maxChars = Math.floor(maxWidth / avgCharWidth);
    return text.length > maxChars ? text.substring(0, maxChars - 3) + '...' : text;
  }

  buildTimelineFromAlerts(alerts: UtmAlertType[]): TimelineItem[] {
    return alerts.map(cha => ({
      startDate: cha['@timestamp'],
      name: cha.name,
      metadata: cha,
      iconUrl: 'assets/icons/echoes/echoes_default.png'
    }));
  }

  private assignYOffsetToGroups(groups: TimelineGroup[]): TimelineGroup[] {
    const FIVE_MINUTE = 5 * 60 * 1000;
    const sorted = [...groups].sort((a, b) => a.startTimestamp - b.startTimestamp);

    for (let i = 0; i < sorted.length; i++) {
      const group = sorted[i];
      group.yOffset = 0; // default base

      for (let j = 0; j < i; j++) {
        const prev = sorted[j];
        const dx = group.startTimestamp - prev.startTimestamp;
        if (dx < FIVE_MINUTE) {
          group.yOffset = Math.max(group.yOffset, (prev.yOffset || 0) + 1);
        }
      }
    }

    return sorted;
  }

  generateTimelineGroups(alerts: UtmAlertType[], intervalMs: number): TimelineGroup[] {
    const items = this.buildTimelineFromAlerts(alerts);
    const groups = this.groupByInterval(items, intervalMs);
    return this.assignYOffsetToGroups(groups);
  }

}
