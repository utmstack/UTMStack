import {getCurrentDateTimeString} from './date.util';

export function replaceBreakLine(doc: string): string {
  if (doc) {
    let msg = doc.split('\n').join('<br>');
    msg = String(msg).split('\t\t').join('&nbsp;');
    msg = String(msg).split('\t').join('&nbsp;&nbsp;');
    return msg;
  } else {
    return doc;
  }
}

export function normalizeString(str: string): string {
  str = str.replace(/\W+(?!$)/g, '-').toLowerCase();
  str = str.replace(/\W$/, '').toLowerCase();
  return str;
}


export const getElementPrefix = (inputString): string => {
  const regex = /\b(IR(?:A)?-\d+)\s/g;
  return inputString.match(regex);
};

export const createElementPrefix = (prefix: string): string => {
  return `${prefix}-${getCurrentDateTimeString()} `;
};
