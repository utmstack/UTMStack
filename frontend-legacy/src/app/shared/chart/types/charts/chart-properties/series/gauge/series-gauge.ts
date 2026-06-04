import {GaugeAxisLine} from './gauge-axis-line';
import {GaugeAxisTick} from './gauge-axis-tick';
import {GaugeDetail} from './gauge-detail';
import {GaugePointer} from './gauge-pointer';

export class SeriesGauge {
  type?: 'gauge';
  startAngle?: number;
  endAngle?: number;
  min?: number;
  max?: number;
  radius?: string;
  center?: string[];
  splitNumber?: number;
  axisLine?: GaugeAxisLine;
  axisTick?: GaugeAxisTick;
  pointer?: GaugePointer;
  detail?: GaugeDetail;
  data?: number[] | { name: string, value: number }[];
  title: {
    offsetCenter: number[],
    color: string
  };
  splitLine?: {           // 分隔线
    length: number,         // 属性length控制线长
    lineStyle: {       // 属性lineStyle（详见lineStyle）控制线条样式
      color: 'auto'
    }
  };


  constructor(type: 'gauge',
              startAngle: number,
              endAngle: number,
              min: number,
              max: number,
              radius: string,
              center: string[],
              splitNumber: number,
              axisLine: GaugeAxisLine,
              axisTick: GaugeAxisTick,
              pointer: GaugePointer,
              detail: GaugeDetail,
              data: number[] | { name: string, value: number }[],
              title: {
                offsetCenter: number[],
                color: string
              },
              splitLine?: {           // 分隔线
                length: number,         // 属性length控制线长
                lineStyle: {       // 属性lineStyle（详见lineStyle）控制线条样式
                  color: 'auto'
                }
              }) {
    this.type = type;
    this.startAngle = startAngle;
    this.endAngle = endAngle;
    this.min = min;
    this.max = max;
    this.radius = radius;
    this.center = center;
    this.splitNumber = splitNumber;
    this.axisLine = axisLine;
    this.axisTick = axisTick;
    this.pointer = pointer;
    this.detail = detail;
    this.data = data;
    this.title = title;
    this.splitLine = splitLine;
  }
}
