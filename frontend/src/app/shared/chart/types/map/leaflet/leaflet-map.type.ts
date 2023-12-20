import {LeafletTilesType} from './leaflet-tiles.type';

export class LeafletMapType {
  center?: [number, number];
  zoom?: number;
  roam?: boolean;
  layerControl?: {
    position: string
  };
  tiles?: LeafletTilesType[];
}
