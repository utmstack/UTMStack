import {Component, Input, OnInit} from '@angular/core';

@Component({
  selector: 'app-utm-country-flag',
  templateUrl: './utm-country-flag.component.html',
  styleUrls: ['./utm-country-flag.component.scss']
})
export class UtmCountryFlagComponent implements OnInit {
  /**
   * Country code
   */
  @Input() countryCode: string;
  /**
   * Country name
   */
  @Input() country: string;

  /**
   * Determine if render country label
   */
  @Input() showCountry: boolean;

  constructor() {
  }

  ngOnInit() {
  }

}
