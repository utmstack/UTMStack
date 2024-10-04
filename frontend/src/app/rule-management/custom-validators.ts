import {AbstractControl, FormArray, ValidationErrors, ValidatorFn} from "@angular/forms";

export const variableTemplate = {get: [], as: [''], of_type: []};
export function urlValidator(control: AbstractControl): ValidationErrors | null {
  const urlPattern = /^(https?|ftp):\/\/[^\s/$.?#].\S*$/i;
  return urlPattern.test(control.value) ? null : {invalidUrl: true};
}

export function singleTermValidator(): ValidatorFn {
  return (control: AbstractControl): ValidationErrors | null => {
    const value = control.value || '';
    const isValid = value.trim().length > 0 && !/\s/.test(value);
    return isValid ? null : {noSpaces: true};
  };
}

export function minLengthArray(min: number) {
  return (control: FormArray): { [key: string]: boolean } | null => {
    if (control.length >= min) {
      return null;
    }
    return {minLengthArray: true};
  };
}
