import {PageableDTO, SortDTO} from '../../../shared/types/page/page.model';

export interface NotificationDTO {
  id: number;
  source: string;
  type: string;
  message: string;
  createdAt: string;
  updatedAt: string | null;
  read: boolean;
}

export interface UtmNotification {
  content?: NotificationDTO[];
  pageable?: PageableDTO;
  last: boolean;
  totalElements: number;
  totalPages: number;
  first: boolean;
  size: number;
  number: number;
  sort: SortDTO;
  numberOfElements: number;
  empty: boolean;
}

