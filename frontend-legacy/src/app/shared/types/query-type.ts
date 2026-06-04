import {Operator} from '../enums/operator.enum';

export class QueryType {
  private query: string;

  /**
   * Build a query based on url parameters
   * @param field : Name of the param
   * @param value : Value of the param
   * @param operator : Operator between param and value
   */
  add(field: string, value: any, operator?: Operator): QueryType {
    if (field && value !== null && typeof value !== 'undefined') {
      if (operator) {
        if (!this.query) {
          this.query = `?${field}.${operator}=${value}`;
        } else {
          this.query += `&${field}.${operator}=${value}`;
        }
      } else {
        if (!this.query) {
          this.query = `?${field}=${value}`;
        } else {
          this.query += `&${field}=${value}`;
        }
      }
    }
    return this;
  }

  /**
   *
   */
  toString(): string {
    return this.query ? this.query : '';
  }
}
