/**
 * Mixing Two colors
 * @param colorOne Color one
 * @param colorTwo Color two
 * @param weight Weight
 */
function mixColor(colorOne, colorTwo, weight): string {
  function d2h(d) {
    return d.toString(16);
  }  // convert a decimal value to hex
  function h2d(h) {
    return parseInt(h, 16);
  }

  weight = (typeof (weight) !== 'undefined') ? weight : 50; // set the weight to 50%, if that argument is omitted

  let color = '#';

  for (let i = 0; i <= 5; i += 2) { // loop through each of the 3 hex pairs—red, green, and blue
    const v1 = h2d(colorOne.substr(i, 2)); // extract the current pairs
    const v2 = h2d(colorTwo.substr(i, 2));

    // combine the current pairs from each source color, according to the specified weight
    let val = d2h(Math.floor(v2 + (v1 - v2) * (weight / 100.0)));

    while (val.length < 2) {
      val = '0' + val;
    } // prepend a '0' if val results in a single digit
    color += val; // concatenate val to our new color string
  }
  return color;
}

function  rgb2hex(rgb) {
  // A very ugly regex that parses a string such as 'rgb(191, 0, 46)' and produces an array
  rgb = rgb.match(/^rgba?\((\d+),\s*(\d+),\s*(\d+)(?:,\s*([\d\.]+))?\)$/);
  function hex(x) {
    // tslint:disable-next-line:radix
    return ('0' + parseInt(x).toString(16)).slice(-2);
  } // another way to convert a decimal to hex

  return (hex(rgb[1]) + hex(rgb[2]) + hex(rgb[3])).toUpperCase(); // concatenate the pairs and return them upper cased
}

export function calcMultipleColors(colors: string[]): string {
  let colorBase = colors[0];
  for (const color of colors) {
    if (color) {
      colorBase = colorBase.replace(/#/g , '');
      const colorTag = color.replace(/#/g , '');
      colorBase = mixColor(colorBase, colorTag, null);
    }
  }
  return colorBase;
}


