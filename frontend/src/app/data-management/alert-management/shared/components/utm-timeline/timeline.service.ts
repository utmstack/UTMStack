import { Injectable } from '@angular/core';
import {TimelineItem} from '../../../../../shared/types/utm-timeline-item';

export interface TimelineGroup {
  startTimestamp: number;
  items: TimelineItem[];
  yOffset?: number;
}

@Injectable()
export class TimelineService {

  /**
   * Groups timeline items by a fixed time interval (milliseconds)
   */
  groupByInterval(items: TimelineItem[], intervalMs: number): TimelineGroup[] {
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

  /**
   * Simple pagination of timeline groups
   */
  paginate(groups: TimelineGroup[], page: number, pageSize: number): TimelineGroup[] {
    const start = page * pageSize;
    return groups.slice(start, start + pageSize);
  }
}
