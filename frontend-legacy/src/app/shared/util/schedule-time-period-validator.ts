import {FormGroup} from '@angular/forms';

/**
 * Openvas Schedule form validator
 * @param group FormGroup
 */
export function timePeriodValidator(group: FormGroup) {
  if (group) {
    group.get('time').setErrors(null);
    if (group.value.durationTime === group.value.periodTime) {
      if (group.value.duration < group.value.period) {
        group.get('time').setErrors(null);
      } else {
        group.get('time').setErrors({time: true});
      }
    } else if (group.value.periodTime === 'hour' && group.value.durationTime !== 'hour') {
      group.get('time').setErrors({time: true});
    } else if (group.value.periodTime === 'day'
      && (group.value.durationTime === 'week' || group.value.durationTime === 'month')) {
      group.get('time').setErrors({time: true});
    } else if (group.value.periodTime === 'week' && group.value.durationTime === 'month') {
      group.get('time').setErrors({time: true});
    } else {
      group.get('time').setErrors(null);
    }
  }
}
