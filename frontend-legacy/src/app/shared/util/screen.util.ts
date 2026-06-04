export function calcTableDimension(pageWidth: number): { filterWidth: number, tableWidth: number } {

  let filterWidth = 0;
  let tableWidth = 0;
  if (pageWidth > 1980) {
    filterWidth = 350;
    tableWidth = pageWidth - filterWidth - 51;
  } else {
    filterWidth = 300;
    tableWidth = pageWidth - filterWidth - 51;
  }
  if (pageWidth > 2500) {
    filterWidth = 350;
    tableWidth = pageWidth - filterWidth - 51;
  }
  if (pageWidth > 4000) {
    filterWidth = 400;
    tableWidth = pageWidth - filterWidth - 51;
  }
  return {filterWidth, tableWidth};
}
