export class ChartModel {
  series: ChartSeriesModel[];
  values: any[];
  legend: any[];
}

export class ChartSeriesModel {
  name: string;
  value?: string | number;
  type?: string;
  smooth?: boolean;
  data?: any[];
  textStyle?: {
    normal?: {
      color?: string
    }
  };
  itemStyle?: {
    color?: string;
    normal?: {
      color?: string
      areaStyle?: {
        type?: string
      };
    }
  };
}
