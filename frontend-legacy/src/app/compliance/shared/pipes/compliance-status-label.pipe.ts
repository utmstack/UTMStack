import { Pipe, PipeTransform } from '@angular/core';
import {getComplianceStatusLabel} from '../enums/compliance-status.enum';

@Pipe({
  name: 'complianceStatusLabel'
})
export class ComplianceStatusLabelPipe implements PipeTransform {
  transform(value: any): string {
    return getComplianceStatusLabel(value);
  }
}
