import {Component, OnInit} from '@angular/core';
import {getDateInfo} from '../../../../../../constants/date-timezone-date.const';
import {TimezoneFormatService} from '../../../../../../services/utm-timezone.service';
import {DatePipeDefaultOptions} from '../../../../../../types/date-pipe-default-options';

@Component({
  selector: 'app-utm-date-format-info',
  templateUrl: './utm-date-format-info.component.html',
  styleUrls: ['./utm-date-format-info.component.scss']
})
export class UtmDateFormatInfoComponent implements OnInit {
  dateFormat: DatePipeDefaultOptions;
  date: { label: string; format: string; equivalentTo: string };

  constructor(private timezoneFormatService: TimezoneFormatService) {
  }

  ngOnInit() {
    this.timezoneFormatService.getDateFormatSubject().subscribe(format => {
      this.dateFormat = format;
      this.date = getDateInfo(format.dateFormat);
    });
  }

}
