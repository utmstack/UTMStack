import {Injectable} from '@angular/core';
import {ElasticDataService} from '../../../../shared/services/elasticsearch/elastic-data.service';
import {DataNatureTypeEnum} from '../../../../shared/enums/nature-data.enum';
import {ElasticOperatorsEnum} from '../../../../shared/enums/elastic-operators.enum';

/**
 * #2937: shared helper for merging an alert's echo-children events into its own
 * events. Used by both the full-detail page and the list drawer so the merge
 * logic lives in one place.
 */
@Injectable({
  providedIn: 'root'
})
export class AlertEchoEventsService {
  constructor(private elasticDataService: ElasticDataService) {
  }

  /**
   * Load echo alerts for `parentId` and merge their events into `parent.events`
   * (de-duplicated on event id). Mutates `parent.events` in place and returns
   * the merged array. Resolves the callback with the merged array even when the
   * lookup fails, so the caller still renders the parent's own events.
   *
   * Fetches up to MAX_ECHO_CHILDREN in one page; the backend `top` caps the
   * result set. Realistic echo counts are well under this (echo rules cap at a
   * few hundred); a single oversized alert truncates to the cap rather than
   * undercounting silently.
   */
  private readonly MAX_ECHO_CHILDREN = 500;

  mergeEchoEvents(parent: any, done: (merged: any[]) => void): void {
    const existing: any[] = (parent && Array.isArray(parent.events)) ? parent.events : [];
    if (!parent || !parent.id) {
      done(existing);
      return;
    }
    this.elasticDataService.search(1, this.MAX_ECHO_CHILDREN, this.MAX_ECHO_CHILDREN, DataNatureTypeEnum.ALERT, [{
      field: 'parentId',
      operator: ElasticOperatorsEnum.IS,
      value: parent.id
    }]).subscribe({
      next: response => {
        const children = (response && response.body) || [];
        const merged = [...existing];
        children.forEach(child => {
          const evs = child && Array.isArray(child.events) ? child.events : [];
          evs.forEach(ev => {
            if (ev && ev.id && merged.some(m => m && m.id === ev.id)) return;
            merged.push(ev);
          });
        });
        if (parent) {
          parent.events = merged;
        }
        done(merged);
      },
      error: () => {
        // If the children lookup fails, the parent's own events still render.
        done(existing);
      }
    });
  }
}
