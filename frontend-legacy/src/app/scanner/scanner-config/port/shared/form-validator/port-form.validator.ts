import {AbstractControl, ValidatorFn} from '@angular/forms';

export function portValidator(): ValidatorFn {
  return (control: AbstractControl): { [key: string]: any } | null => {
    const port: string[] = control.value.toString().split(',');
    let valid = false;
    port.forEach(p => {
      if (p.includes('-')) {
        const range = p.split('-');
        valid = Number.parseInt(range[0], 0) > Number.parseInt(range[1], 0);
      } else {
        const num = Number.parseInt(p, 0);
        if (isNaN(num)) {
          valid = true;
        } else {
          valid = typeof num !== 'number';
        }
      }
    });
    if (valid) {
      return {pattern: {valid}};
    } else {
      return null;
    }
  };
}

