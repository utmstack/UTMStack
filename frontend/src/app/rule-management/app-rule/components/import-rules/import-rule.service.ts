import {Injectable} from '@angular/core';
import {Rule} from '../../../models/rule.model';

@Injectable()
export class ImportRuleService {

   isValidURL(url: string): boolean {
    try {
      new URL(url);
      return true;
    } catch {
      return false;
    }
  }

  isValidRule(obj: Rule): obj is Rule {
     console.log('Valid');
    if (!obj || typeof obj !== 'object') { return false; }

     const requiredProps: (keyof Rule)[] = [
       'dataTypes', 'name', 'confidentiality', 'integrity', 'availability',
      'category', 'technique', 'description', 'references', 'definition',
    ];


    if (!requiredProps.every(prop => prop in obj)) { return false; }

     if (
      !Array.isArray(obj.dataTypes) ||
      typeof obj.name !== 'string' ||
      typeof obj.confidentiality !== 'number' ||
      typeof obj.integrity !== 'number' ||
      typeof obj.availability !== 'number' ||
      typeof obj.category !== 'string' ||
      typeof obj.technique !== 'string' ||
      typeof obj.description !== 'string' ||
      !Array.isArray(obj.references) ||
      typeof obj.definition !== 'object'
    ) {
      return false;
    }

     return obj.references.every((ref: any) => typeof ref === 'string' && this.isValidURL(ref));

  }
}
