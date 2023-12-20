export interface MultilineSerie {
  name: string;
  type: 'line';
  smooth: true;
  data: number[];
  symbolSize?: 6;
  areaStyle?: {
    normal?: {
      opacity?: 0.25
    }
  };
  itemStyle?: any;
}
