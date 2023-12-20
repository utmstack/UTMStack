import * as flat from 'flat';
import {UtmFieldType} from '../types/table/utm-field.type';

/**
 *
 * @param obj Object to find value
 * @param path Path to property
 * @param def Default value if not exist
 */
export function getValueFromPropertyPath(obj: object, path: string, def: any = null) {
  // Get the path as an array
  path = stringToPath(path);
  // Cache the current object
  let current = obj;
  // For each item in the path, dig into the object
  // tslint:disable-next-line:prefer-for-of
  for (let i = 0; i < path.length; i++) {

    // If the item isn't found, return the default (or null)
    if (!current[path[i]]) {
      return def;
    }
    // Otherwise, update the current  value
    current = current[path[i]];

  }
  return current;
}

/**
 * /**
 * If the path is a string, convert it to an array
 * @return The path array
 * @param path Property to return value
 */
export function stringToPath(path) {
  // If the path isn't a string, return it
  if (typeof path !== 'string') {
    return path;
  }
  // Create new array
  const output = [];
  // Split to an array with dot notation
  path.split('.').forEach((item) => {
    // Split to an array with bracket notation
    item.split(/\[([^}]+)\]/g).forEach((key) => {

      // Push to the new array
      if (key.length > 0) {
        output.push(key);
      }

    });

  });
  return output;
}

/**
 * Return value from object based on path to value
 * @param row Object
 * @param field Path to get data
 */
export function extractValueFromObjectByPath(row: any, field: UtmFieldType) {
  const objectRow = flat(row);
  const fieldExtract = field.field.includes('.keyword') ?
    field.field.replace('.keyword', '') : field.field;
  const tdValue = convertByType(objectRow[fieldExtract]);
  return tdValue ? tdValue : '-';
}

export function convertByType(data: any) {

  if (data && typeof data === 'object') {
    return JSON.stringify(data);
  } else if (typeof data === 'boolean') {
    return data ? 'Yes' : 'No';
  } else {
    return data;
  }
}

export function convertObjectToKeyValueArray(object) {
  const target: { key: string, value: any }[] = [];
  if (object) {
    for (const [key, value] of Object.entries(flat(object))) {
      target.push({key, value});
    }
    return target;
  }
}

export function findInObject(obj, item) {
  for (const key in obj) {
    if (obj[key] && typeof obj[key] === 'object') {
      const result = findInObject(obj[key], item);
      if (result) {
        result.unshift(key);
        return result;
      }
    } else if (obj[key] === item) {
      return [key];
    }
  }
}



