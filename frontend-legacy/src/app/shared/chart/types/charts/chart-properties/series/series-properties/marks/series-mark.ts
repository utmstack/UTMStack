import {MarkData} from './mark-data';

export class SeriesMark {
  symbol?: 'circle' | 'rect' | 'roundRect' | 'triangle' | 'diamond' | 'pin' | 'arrow' | 'none' | string;
  symbolSize?: number;
  data?: MarkData[];
  label?: { fontSize: number };

  constructor(symbol?: 'circle' | 'rect' | 'roundRect' | 'triangle'
    | 'diamond' | 'pin' | 'arrow' | 'none' | string,
              symbolSize?: number,
              data?: MarkData[],
              label?: { fontSize: number }) {
    this.symbol = symbol ? symbol : 'pin';
    this.symbolSize = symbolSize ? symbolSize : null;
    this.data = data ? data : [new MarkData()];
    this.label = label ? label : {fontSize: 10};
  }
}
