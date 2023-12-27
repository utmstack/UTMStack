export class EchartClickAction {
  // type of the component to which the clicked glyph belongs
  // i.e., 'series', 'markLine', 'markPoint', 'timeLine'
  componentType?: string;
  // series type (make sense when componentType is 'series')
  // i.e., 'line', 'bar', 'pie'
  seriesType?: string;
  // series index in incoming option.series (make sense when componentType is 'series')
  seriesIndex?: number;
  // series name (make sense when componentType is 'series')
  seriesName?: string;
  // data name, category name
  name?: string;
  // data index in incoming data array
  dataIndex?: number;
  // incoming rwa data item
  data?: any | object;
  // Some series, such as sankey or graph, maintains more than
  // one type of data (nodeData and edgeData), which can be
  // distinguished from each other by dataType with its value
  // 'node' and 'edge'.
  // On the other hand, most series has only one type of data,
  // where dataType is not needed.
  dataType?: string;
  // incoming data value
  value?: number | Array<any>;
  // color of component (make sense when componentType is 'series')
  color?: string;
  // User info (only available in graphic component
  // and custom series, if element option has info
  // property, e.g., {type?: 'circle', info?: {some?: 123}})
  info?: any;
}
