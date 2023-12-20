import {Component, Input, OnDestroy, OnInit} from '@angular/core';
import * as L from 'leaflet';
import {MAP_TILE} from '../../../../../shared/constants/map.const';
import {ChartTypeEnum} from '../../../../../shared/enums/chart-type.enum';
import {AlertLocationType} from '../../../../../shared/types/alert/alert-location.type';
import {UtmAlertType} from '../../../../../shared/types/alert/utm-alert.type';
import {calculateMidpoint, getZoomLevel} from '../../../../../shared/util/map.util';
import {getLocationFromAlert} from '../../util/alert-util-function';

@Component({
  selector: 'app-alert-map-location',
  templateUrl: './alert-map-location.component.html',
  styleUrls: ['./alert-map-location.component.scss']
})
export class AlertMapLocationComponent implements OnInit, OnDestroy {
  map: any;
  chartType = ChartTypeEnum;
  @Input() alert: UtmAlertType;
  alertLocations: AlertLocationType[] = [];

  constructor() {
  }

  ngOnInit() {
    getLocationFromAlert(this.alert).then(locations => {
      this.map = new L.map('mapLocation', {
        minZoom: 1,
      });
      const midpoint = calculateMidpoint(this.alert.source.coordinates, this.alert.destination.coordinates);
      const zoomLevel = getZoomLevel(this.alert.source.coordinates, this.alert.destination.coordinates);
      this.map.setView(midpoint, zoomLevel);
      this.map.panTo(midpoint);
      this.alertLocations = locations;
      if (this.alertLocations.length > 0) {
        L.tileLayer(MAP_TILE, {
          attribution: '',
          maxZoom: 18
        }).addTo(this.map);
        if (this.alertLocations.length > 1) {
          const latlngs = [
            this.alertLocations[0].location,
            this.alertLocations[1].location
          ];
          const polyline = L.polyline(latlngs, {
            color: '#5c6bc0',
            width: .2,
            type: 'dotted',
            dashOffset: '15px'
          }).addTo(this.map);

          this.map.fitBounds(polyline.getBounds());

          L.circle(this.alertLocations[0].location, this.alertLocations[0].accuracy, {
            color: '#c05050',
            fillColor: 'rgba(192,80,80,0.61)',
            fillOpacity: 0.8,
            weight: 1
          }).addTo(this.map);
          L.circle(this.alertLocations[1].location, this.alertLocations[1].accuracy, {
            color: '#004b8b',
            fillColor: 'rgba(0,75,139,0.71)',
            fillOpacity: 0.8,
            weight: 1
          }).addTo(this.map);
        }
        for (const dat of this.alertLocations) {
          const tooltip = '<div style="' +
              'white-space: pre-line;' +
              'padding-top:10px' +
              'font: 13px / 20px Poppins, sans-serif;' +
              'pointer-events: none;">' +
              '<span style="display:inline-block;' +
              'margin-right:5px;' +
              'border-radius:10px;' +
              'width:10px;' +
              'height:10px;' +
              'background-color:' + (dat.locationType === 'source' ? '#c05050' : '#004b8b') + ';">' +
              '</span>' + (dat.locationType.charAt(0).toUpperCase() + dat.locationType.substring(1, dat.locationType.length)) +
              ':<span class="mt-1 font-weight-semibold"> ' + dat.ip + '</span></div>';
          const myIcon = new L.Icon({
            iconUrl: dat.locationType === 'source' ? '/assets/img/red_marker.png' : '/assets/img/blue_marker.png',
            iconSize: [15, 25],
            iconAnchor: [4, 25],
            popupAnchor: [4, -30],
          });
          const marker = L.marker(dat.location, {icon: myIcon})
            .bindPopup(tooltip, {className: 'class-popup'});
          // markers.addLayer(marker);
          marker.addTo(this.map);
          marker.on('mouseover', (ev) => {
            marker.openPopup();
          });
          marker.on('mouseout', (ev) => {
            marker.closePopup();
          });
        }
      }
    });
    this.map.invalidateSize();
  }

  ngOnDestroy(): void {
    document.getElementById('mapLocation').remove();
  }

  hasCoord(type: string) {
    return this.alertLocations.findIndex(value => value.locationType === type) !== -1;
  }
}
