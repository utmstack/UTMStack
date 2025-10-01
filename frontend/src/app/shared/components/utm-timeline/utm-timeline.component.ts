import { Component, Input, OnChanges, SimpleChanges, Output, EventEmitter } from '@angular/core';
import {TimelineItem} from '../../types/utm-timeline-item'


@Component({
  selector: 'utm-timeline',
  templateUrl: './utm-timeline.component.html',
  styleUrls: ['./utm-timeline.component.scss']
})
export class UtmTimeLineComponent implements OnChanges {
  @Input() items: TimelineItem[] = [];
  @Input() title: string = '';
  @Output() itemClick = new EventEmitter<TimelineItem>();

  chartOption: any = {};

  ngOnChanges(changes: SimpleChanges): void {
    if ((this.items.length || 0) > 0) {
      this.buildChart();
    }
  }

  private buildChart() {
    const groups: Record<number, TimelineItem[]> = {};
    this.items.forEach(item => {
      const ts = new Date(item.startDate).getTime();
      if (!groups[ts]) groups[ts] = [];
      groups[ts].push(item);
    });

    const seriesData: any[] = [];
    Object.keys(groups).forEach(tsStr => {
      const ts = Number(tsStr);
      const group = groups[ts];

      group.forEach((item, idx) => {
        seriesData.push({
          value: [
            ts,                                              // index 0: timestamp
            0,                                               // index 1: Y coordinate
            item.name,                                       // index 2: name
            new Date(item.startDate).toLocaleString(),      //  index 3:formatted date
            !!item.iconUrl?item.iconUrl:'',                 //  index 4: offsetIndex
            idx,                                             // index 5: offsetIndex
            group.length                                     // index 6: groupSize
          ],
          itemData: item
        });
      });
    });

    this.chartOption = {
      title: {
        text: this.title,
        left: 'center',
        textStyle: {
          fontSize: 16,
          fontWeight: 'bold'
        }
      },
      grid:{
        left:0,
        right:0
      },
      tooltip: {
        trigger: 'item',
        formatter: (params: any) =>
          `<b>${params.data.value[2]}</b><br/>${params.data.value[3]}`
      },
      xAxis: {
        type: 'time',
        //let a space between init of the timeline and the first/last element so they can be showed
       min: (value: any) => {
          return value.min - 5 * 1000;
        },
        max: (value: any) => {
          return value.max + 10 * 1000;
        },
        axisLabel: {
          formatter: (val: number) => {
            const date = new Date(val)
            return `${date.getHours()%12}:${date.getMinutes() < 10 ? '0' : ''}${date.getMinutes()}:${date.getSeconds() < 10 ? '0' : ''}${date.getSeconds()} \n ${date.getDate()}/${date.getMonth() + 1}/${date.getFullYear()}  `
          }
        },
      },
      yAxis: { show: false, type: 'value'},
      series: [
        {
          type: 'custom',
          data: seriesData,
          renderItem: (params: any, api: any) => {
            const timestamp = api.value(0);
            const coord = api.coord([timestamp, 0]);

            const cardHeight = 60;
            const spacing = 10;
            const offsetIndex = api.value(5) || 0;
            const totalOffset = (cardHeight + spacing) * offsetIndex + 80;

            const name = api.value(2);
            const date = api.value(3);
            const iconUrl = api.value(4);

            const iconPadding = 0;

            return {
              type: 'group',
              children: [
                {
                  type: 'rect',
                  shape: {
                    x: coord[0] - 75,
                    y: coord[1] - totalOffset - cardHeight,
                    width: 250,
                    height: cardHeight,
                    r: 8
                  },
                  style: {
                    fill: '#fefefe',
                    stroke: '#0277bd',
                    lineWidth: 1,
                    shadowBlur: 8,
                    shadowColor: 'rgba(0,0,0,0.2)',
                    cursor: 'pointer',
                    overflow:'hidden'
                  }
                },
                {
                  type: 'image',
                  style: {
                    image: iconUrl,
                    x: coord[0] - 75 + iconPadding +1.5,
                    y: coord[1] - totalOffset - cardHeight+1.5,
                    width: cardHeight-2.5,
                    height:cardHeight-2.5,
                    r: 8
                  }
                },
                {
                  type: 'text',
                  style: {
                    x: coord[0] - 75 + iconPadding + cardHeight + 10,
                    y: coord[1] - totalOffset - cardHeight + 20,
                    text:this.truncateText(name,150),
                    textAlign: 'left',
                    fill: '#000',
                    fontSize: 13,
                    fontWeight: 600,
                    width: 150 - cardHeight - iconPadding * 2 - 10,
                    overflow: 'break',
                    ellipsis: '...'
                  }
                },
                {
                  type: 'text',
                  style: {
                    x: coord[0] - 75 + iconPadding + cardHeight + 10,
                    y: coord[1] - totalOffset - cardHeight + 40,
                    text: date,
                    textAlign: 'left',
                    fill: '#666',
                    fontSize: 10
                  }
                },
                {
                  type: 'line',
                  shape: {
                    x1: coord[0],
                    y1: coord[1] - totalOffset,
                    x2: coord[0],
                    y2: coord[1]
                  },
                  style: { stroke: '#0277bd', lineWidth: 1.5 }
                }
              ]
            };
          },
          encode: { x: 0, y: 1 }
        }
      ],
      dataZoom: [
        { type: 'slider', xAxisIndex: 0 },
        { type: 'inside', xAxisIndex: 0 }
      ]
    };
  }

  onChartClick(event: any): void {
    if (event.data && event.data.itemData) {
      this.itemClick.emit(event.data.itemData.metadata);
    }
  }

 truncateText(text: string, maxWidth: number) {
  const avgCharWidth = 7;
  const maxChars = Math.floor(maxWidth / avgCharWidth);
  return text.length > maxChars ? text.substring(0, maxChars - 3) + '...' : text;
};

}
