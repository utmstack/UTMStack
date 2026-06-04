import { AbstractControl, ValidationErrors, ValidatorFn } from '@angular/forms';

export function containsVariable(variables: string[]): ValidatorFn {
  return (control: AbstractControl): ValidationErrors | null => {
    const value = control.value || '';
    const hasVariable = variables.some(v => value.includes(v));
    return hasVariable ? null : { noVariableUsed: true };
  };
}

export function minWordsValidator(minWords: number, minLengthPerWord: number): ValidatorFn {
  return (control: AbstractControl): ValidationErrors | null => {
    const value = control.value.trim();

    if (!value) { return { minWords: { requiredWords: minWords, minLengthPerWord } }; }

    const words = value
      .split(/\s+/)
      .filter(word => word.length >= minLengthPerWord);

    return words.length >= minWords ? null : {
      minWords: { requiredWords: minWords, minLengthPerWord }
    };
  };
}


export function uniqueStringValidator(existing: string[]): ValidatorFn {
  return (control: AbstractControl): ValidationErrors | null => {
    if (!control.value) { return null; }

    const normalized = control.value.trim().toLowerCase();
    const exists = existing
      .map(x => x.trim().toLowerCase())
      .includes(normalized);

    return exists ? { unique: true } : null;
  };
}



