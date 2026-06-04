export interface TimelineItem {
  startDate: string | Date;
  name: string;
  metadata: any;
  iconUrl: string | undefined | null;
  yOffset?: number;
}
