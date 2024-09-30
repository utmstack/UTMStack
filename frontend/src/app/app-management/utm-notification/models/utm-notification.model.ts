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

export interface NotificationRequest {
    page: number;
    size: number;
    sort: string;
    source?: string;
    type?: string;
    from?: string;
    to?: string;
}

export enum NotificationSource {
  AS400,
  BIT_DEFENDER,
  AWS,
  AZURE,
  OFFICE_365,

  SOPHOS,
  GOOGLE,
  EMAIL_SETTING,
  ALERTS
}

export enum NotificationStatus {
  ACTIVE,
  HIDDEN,
  DELETED
}

