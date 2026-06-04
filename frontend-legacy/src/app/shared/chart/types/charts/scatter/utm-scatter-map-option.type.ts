import {LeafletMapType} from '../../map/leaflet/leaflet-map.type';
import {SeriesScatter} from '../chart-properties/series/scatter/series-scatter';
import {Tooltip} from '../chart-properties/tooltip/tooltip';
import {VisualMap} from '../chart-properties/visualmap/visual-map';

export class UtmScatterMapOptionType {
  tooltip?: Tooltip;
  visualMap?: VisualMap;
  leaflet?: LeafletMapType;
  top?: number;
  series?: SeriesScatter[];
}
