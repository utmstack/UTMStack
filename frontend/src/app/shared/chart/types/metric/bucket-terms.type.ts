export class BucketTermsType {
  sortBy: string;
  asc: boolean;
  size: number;

  constructor(sortBy?: string, asc?: boolean, size?: number) {
    this.sortBy = sortBy ? sortBy : '_count';
    this.asc = asc ? asc : false;
    this.size = size ? size : 5;
  }
}
