export class GaugeAxisLine {
  show: boolean;
  lineStyle: {
    color?: [number, string][],
    width?: number
  };


  constructor(show: boolean, lineStyle: { color: [number, string][]; width: number }) {
    this.show = show;
    this.lineStyle = lineStyle ? lineStyle : {
      color:
        [
          [.3, 'red'],
          [.7, '#ff4500'],
          [1, 'lightgreen'],
        ],
      width: 12
    };
  }
}
