import {SortDirection} from './sort-direction.type';

export interface SortEvent {
  column: string;
  direction: SortDirection;
}
