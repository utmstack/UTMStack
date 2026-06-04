export interface SortDTO {
  empty: boolean;
  unsorted: boolean;
  sorted: boolean;
}

export interface PageableDTO {
  sort?: SortDTO;
  offset: number;
  pageNumber: number;
  pageSize: number;
  unpaged: boolean;
  paged: boolean;
}
