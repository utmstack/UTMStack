import {HttpParams} from '@angular/common/http';

/**
 * Append params to query in request
 * @param req Any object
 */
export const createRequestOption = (req?: any): HttpParams => {
  let options: HttpParams = new HttpParams();

  const appendParam = (key: string, value: any) => {
    if (value !== undefined && value !== null && value !== '') {
      options = options.set(key, value);
    }
  };

  if (req) {
    Object.keys(req).forEach(key => {
      const value = req[key];
      if (value !== undefined && value !== null) {
        if (key === 'sort' && Array.isArray(value)) {
          value.forEach((val: string) => {
            options = options.append('sort', val);
          });
        } else if (typeof value === 'object' && !Array.isArray(value)) {
          Object.entries(value).forEach(([nestedKey, nestedValue]) => {
            appendParam(`${key}[${nestedKey}]`, nestedValue);
          });
        } else {
          appendParam(key, value);
        }
      }
    });
  }

  return options;
};

