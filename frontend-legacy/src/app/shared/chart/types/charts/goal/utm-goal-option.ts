export class UtmGoalOption {
  metricId?: number;
  value?: number;
  append?: string;
  max?: number;
  animate?: boolean;
  min?: number;
  thick?: number;
  cap?: 'round' | 'butt';
  type?: 'full' | 'arch' | 'semi';
  thresholds?: object;
  label?: string;
  foregroundColor?: string;
  decimal: number;

  constructor(metricId?: number,
              value?: number,
              append?: string,
              max?: number,
              animate?: boolean,
              min?: number,
              thick?: number,
              cap?: 'round' | 'butt',
              type?: 'full' | 'arch' | 'semi',
              thresholds?: object,
              label?: string,
              foregroundColor?: string) {
    this.metricId = metricId;
    this.value = value;
    this.append = append;
    this.max = max;
    this.animate = animate;
    this.min = min;
    this.thick = thick;
    this.cap = cap;
    this.type = type;
    this.thresholds = thresholds;
    this.label = label;
    this.foregroundColor = foregroundColor;
  }


}
