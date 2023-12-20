import {Injectable} from '@angular/core';
import {AbstractControl} from '@angular/forms';

@Injectable({
    providedIn: 'root'
})
export class InputClassResolve {
    constructor() {
    }

    /**
     * Return class to apply to input base on errors
     * @param control AbstractControl in form
     */
    resolveClassInput(control: AbstractControl): string {
        if (control.errors && (control.touched || control.dirty)) {
            return 'input-field-has-error';
        } else {
            return 'input-field-has-noerror';
        }
    }
}
