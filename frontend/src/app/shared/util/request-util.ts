import {HttpParams} from '@angular/common/http';

/**
 * Append params to query in request
 * @param req Any object
 */
export const createRequestOption = (req?: any): HttpParams => {
  let options: HttpParams = new HttpParams();
  if (req) {
    Object.keys(req).forEach(key => {
      if (key !== 'sort' && req[key] !== undefined && req[key] !== '' && req[key] !== null) {
        options = options.set(key, req[key]);
      }
    });
    if (req.sort) {
      options = options.append('sort', req.sort);
    }
  }
  return options;
};
