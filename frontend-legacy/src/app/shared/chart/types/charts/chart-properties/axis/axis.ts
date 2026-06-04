export class Axis {
  name?: string;
  type?: 'value' | 'category' | 'time' | 'log';
  data?: any[];
  boundaryGap?: boolean;
  axisLabel?: {
    color?: string
    formatter?: string
  };
  axisLine?: {
    lineStyle?: {
      color?: string
    }
  };
  splitLine?: {
    show?: boolean,
    lineStyle?: {
      color?: string,
      type?: 'solid' | 'dashed' | 'dotted'
    }
  };

  constructor(type?: 'value' | 'category' | 'time' | 'log',
              name?: string,
              data?: any[],
              boundaryGap?: boolean,
              axisLabel?: { color?: string },
              axisLine?: { lineStyle?: { color?: string } },
              splitLine?: { show?: boolean; lineStyle?: { color?: string, type?: 'solid' | 'dashed' | 'dotted' } }) {
    this.name = name;
    this.type = type;
    this.data = data;
    this.boundaryGap = boundaryGap ? boundaryGap : false;
    this.axisLabel = axisLabel ? axisLabel : {color: '#333,', formatter: ''};
    this.axisLine = axisLine ? axisLine : {lineStyle: {color: '#999'}};
    this.splitLine = splitLine ? splitLine : {
      show: true,
      lineStyle: {
        color: '#eee',
        type: 'dashed'
      }
    };
  }
}
