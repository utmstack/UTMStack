import { Component, OnInit, ViewChild, ElementRef, OnDestroy } from '@angular/core';
import * as echarts from 'echarts';

@Component({
  selector: 'utm-timeline',
  template: `
    <div class="timeline-wrapper">
      <div #chartContainer class="chart-container"></div>
    </div>
  `,
  styles: [`
    .timeline-wrapper {
      width: 100%;
      height: 100%;
      min-height: 500px;
    }
    .chart-container {
      width: 100%;
      height: 100%;
    }
  `]
})
export class UtmTimeLineComponent implements OnInit, OnDestroy {
  @ViewChild('chartContainer') chartContainer: ElementRef;

  private chart: any;
  public timelineData: any[] = [];

  ngOnInit() {
    // Datos de ejemplo
    this.timelineData = [
      { name: 'Proyecto A', tasks: [
        { task: 'Diseño', start: '2024-01-01', end: '2024-02-15', color: '#5470c6' },
        { task: 'Desarrollo', start: '2024-02-16', end: '2024-05-30', color: '#91cc75' }
      ]},
      { name: 'Proyecto B', tasks: [
        { task: 'Investigación', start: '2024-01-15', end: '2024-03-01', color: '#5470c6' },
        { task: 'Implementación', start: '2024-03-02', end: '2024-06-15', color: '#91cc75' }
      ]},
      { name: 'Proyecto C', tasks: [
        { task: 'Testing', start: '2024-04-01', end: '2024-05-15', color: '#fac858' }
      ]}
    ];
  }

  ngAfterViewInit() {
    setTimeout(() => this.initChart(), 0);
  }

  ngOnDestroy() {
    if (this.chart) {
      this.chart.dispose();
    }
  }

  private initChart() {
    this.chart = echarts.init(this.chartContainer.nativeElement);

    const categories: string[] = [];
    const data: any[] = [];

    // Preparar datos
    this.timelineData.forEach((project, idx) => {
      categories.push(project.name);
      project.tasks.forEach((task: any) => {
        data.push({
          name: task.task,
          value: [
            idx,
            new Date(task.start).getTime(),
            new Date(task.end).getTime(),
            new Date(task.end).getTime() - new Date(task.start).getTime()
          ],
          itemStyle: { color: task.color }
        });
      });
    });

    const option = {
      tooltip: {
        formatter: (params: any) => {
          const start = new Date(params.value[1]);
          const end = new Date(params.value[2]);
          const days = Math.ceil(params.value[3] / (1000 * 60 * 60 * 24));
          return `<b>${params.name}</b><br/>
                  Inicio: ${start.toLocaleDateString()}<br/>
                  Fin: ${end.toLocaleDateString()}<br/>
                  Duración: ${days} días`;
        }
      },
      title: {
        text: 'Timeline de Proyectos',
        left: 'center'
      },
      dataZoom: [{
        type: 'slider',
        filterMode: 'weakFilter',
        height: 20,
        bottom: 20,
        start: 0,
        end: 100,
        handleIcon: 'M10.7,11.9H9.3c-4.9,0.3-8.8,4.4-8.8,9.4c0,5,3.9,9.1,8.8,9.4h1.3c4.9-0.3,8.8-4.4,8.8-9.4C19.5,16.3,15.6,12.2,10.7,11.9z M13.3,24.4H6.7V23h6.6V24.4z M13.3,19.6H6.7v-1.4h6.6V19.6z',
        handleSize: '80%',
        showDetail: false
      }, {
        type: 'inside',
        filterMode: 'weakFilter'
      }],
      grid: {
        top: 60,
        bottom: 80,
        left: 100,
        right: 50
      },
      xAxis: {
        type: 'time',
        axisLabel: {
          formatter: (val: number) => new Date(val).toLocaleDateString('es-ES', { month: 'short', day: 'numeric' })
        }
      },
      yAxis: {
        type: 'category',
        data: categories
      },
      series: [{
        type: 'custom',
        renderItem: (params: any, api: any) => {
          const categoryIndex = api.value(0);
          const start = api.coord([api.value(1), categoryIndex]);
          const end = api.coord([api.value(2), categoryIndex]);
          const height = api.size([0, 1])[1] * 0.6;

          return {
            type: 'rect',
            shape: {
              x: start[0],
              y: start[1] - height / 2,
              width: end[0] - start[0],
              height: height
            },
            style: api.style()
          };
        },
        encode: { x: [1, 2], y: 0 },
        data: data
      }]
    };

    this.chart.setOption(option);
    window.addEventListener('resize', () => this.chart.resize());
  }

  // Métodos públicos
  public setData(data: any[]) {
    this.timelineData = data;
    this.initChart();
  }

  public addProject(project: any) {
    this.timelineData.push(project);
    this.initChart();
  }

  public clear() {
    this.timelineData = [];
    this.initChart();
  }
}

