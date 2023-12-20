export class SeriesTagCloud {
  name?: string;
  type?: 'wordCloud';
  size?: [string, string];
  textRotation?: [number, number, number, number];
  textPadding?: 0;
  autoSize?: {
    enable?: true;
    minSize?: 14
  };
  data?: any[];
  shape?: 'circle' | 'cardioid' | 'diamond' | 'triangle-forward' | 'triangle' | 'star';
  color?: string[];

}
