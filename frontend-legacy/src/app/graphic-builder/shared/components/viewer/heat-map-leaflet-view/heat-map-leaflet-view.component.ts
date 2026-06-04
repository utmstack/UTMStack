import {HttpClient} from '@angular/common/http';
import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import * as L from 'leaflet';
import 'leaflet.heat/dist/leaflet-heat.js';
import {VisualizationType} from '../../../../../shared/chart/types/visualization.type';
import {ChartTypeEnum} from '../../../../../shared/enums/chart-type.enum';

@Component({
  selector: 'app-heat-map-leaflet-view',
  templateUrl: './heat-map-leaflet-view.component.html',
  styleUrls: ['./heat-map-leaflet-view.component.scss']
})
export class HeatMapLeafletViewComponent implements OnInit {
  @Input() building: boolean;
  @Input() chart: ChartTypeEnum;
  @Input() visualization: VisualizationType;
  @Input() width: string;
  @Input() height: string;
  @Output() runned = new EventEmitter<string>();
  loadingOption: boolean;
  data: any[] = [];

  constructor(private httpClient: HttpClient) {
  }

  ngOnInit() {
    const osmLink = '<a href="http://openstreetmap.org">OpenStreetMap</a>';
    const thunLink = '<a href="http://thunderforest.com/">Thunderforest</a>';

    const osmUrl = 'https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png';
    const osmAttrib = '&copy; ' + osmLink + ' Contributors';
    const landUrl = 'https://{s}.tile.thunderforest.com/landscape/{z}/{x}/{y}.png';
    const thunAttrib = '&copy; ' + osmLink + ' Contributors & ' + thunLink;

    const osmMap = L.tileLayer(osmUrl, {attribution: osmAttrib});
    const landMap = L.tileLayer(landUrl, {attribution: thunAttrib});
    // L.control.layers(baseLayers).addTo(map);
    let addressPoints = [];

    for (let i = 0; i < 1000; i++) {
      addressPoints.push([this.getRandomInRange(0, 90, 6),
        this.getRandomInRange(0, -90, 6),
        Math.random()]);
    }

    const map = L.map('map').setView([40.424127, -3.711656], 5.4);
    L.tileLayer('http://{s}.tiles.wmflabs.org/bw-mapnik/{z}/{x}/{y}.png',
      {maxZoom: 19}).addTo(map);

    addressPoints = addressPoints.map((p) => {
      return [p[0], p[1]];
    });
    const heat = L.heatLayer(addressPoints, {
      radius: 25,
      minOpacity: 0.2,
      blur: 15,
    }).addTo(map);
  }

  getRandomInRange(from, to, fixed) {
    return (Math.random() * (to - from) + from).toFixed(fixed) * 1;
    // .toFixed() returns string, so ' * 1' is a trick to convert to number
  }

}
