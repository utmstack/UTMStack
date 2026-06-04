/**
 * Clear object to delete properties where the value is null;
 * @param object Object to clear null values;
 */
export function deleteNullValues(object: any): any {
  if (object) {
    Object.keys(object).forEach(key => {
      if (object[key] === null) {
        delete object[key];
      } else {
        Object.keys(object[key]).forEach(keyChild => {
          if (object[key][keyChild] === null) {
            delete object[key][keyChild];
          } else {
            if (object[key][keyChild] !== undefined) {
              Object.keys(object[key][keyChild]).forEach(lastKey => {
                if (object[key][keyChild][lastKey] === null) {
                  delete object[key][keyChild][lastKey];
                }
              });
            }
          }
        });
      }
    });
  }
  return object;
}
