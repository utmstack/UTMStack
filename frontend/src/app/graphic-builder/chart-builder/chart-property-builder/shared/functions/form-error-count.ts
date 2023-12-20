import {FormArray, FormGroup} from '@angular/forms';

/**
 * Return amount of error from option form
 * @param container FormFormGroup
 */
export function getErrorCountForm(container: FormGroup): number {
  let errorCount = 0;
  for (const controlKey in container.controls) {
    if (container.controls.hasOwnProperty(controlKey)) {
      if (container.controls[controlKey].errors) {
        errorCount += Object.keys(container.controls[controlKey].errors).length;
      }
    }
  }
  return errorCount;
}

/**
 * Return amount of error from form array
 * @param container FormArray
 */
export function getErrorCountFormArray(container: FormArray): number {
  let errorCount = 0;
  for (const controlKey of container.controls) {
    // console.log(container);
    if (controlKey.status === 'INVALID') {
      errorCount += 1;
    }
  }
  return errorCount;
}
