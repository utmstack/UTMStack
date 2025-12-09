import { HttpClient } from '@angular/common/http';
import { AbstractControl, AsyncValidatorFn, ValidationErrors } from '@angular/forms';
import { Observable, of } from 'rxjs';
import { catchError, debounceTime, first, map } from 'rxjs/operators';

export function validateMetadataUrl(http: HttpClient): AsyncValidatorFn {
  return (control: AbstractControl): Observable<ValidationErrors | null> => {
    if (!control.value) {
      return of(null);
    }

    return validateFileUrl(control.value, http).pipe(
      debounceTime(500),
      map(() => null),
      catchError(() => of({ invalidMetadataUrl: true })),
      first()
    );
  };
}

export function validateFileUrl(metadataUrl: string, http: HttpClient): Observable<any> {
  return http.get(metadataUrl, { responseType: 'text' }).pipe(
    map((response: string) => {
      if (!response.includes('EntityDescriptor') && !response.includes('SPSSODescriptor')) {
        throw new Error('Invalid SAML metadata');
      }
      return response;
    })
  );
}

