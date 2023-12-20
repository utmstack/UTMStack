import {VisualizationType} from '../../../../shared/chart/types/visualization.type';

export function cleanVisualizationData(cleanVis: VisualizationType) {
  return new Promise<VisualizationType>(resolve => {
    const visualization: VisualizationType = JSON.parse(JSON.stringify(cleanVis));
    const pathsToClean = ['chartConfig.legend.data', 'chartConfig.data',
      'chartConfig.series', 'chartConfig.xAxis.data', 'chartConfig.yAxis.data'];
    pathsToClean.forEach(path => {
      if (hasPropertyPath(visualization, path)) {
        path.split('.').reduce((o, p, i, arr) => {
          if (i === arr.length - 1) {
            o[p] = [];
          }
          return o[p];
        }, visualization);
      }
    });
    resolve(visualization);
  });
}

function hasPropertyPath(obj, path) {
  return path.split('.').every((prop) => {
    if (!obj || !obj.hasOwnProperty(prop)) {
      return false;
    }
    obj = obj[prop];
    return true;
  });
}
