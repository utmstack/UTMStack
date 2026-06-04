export function getZoomLevel(coord1: number[], coord2: number[]): number {
  // The constant values assume a map of the world that is 256 pixels wide
  const distance: number = getDistance(coord1, coord2);
  const WORLD_PX_HEIGHT = 256;
  const WORLD_PX_WIDTH = 256;
  const ZOOM_MAX = 18;

  // Calculate the zoom level based on the distance
  let zoom = Math.log2(WORLD_PX_WIDTH * 360 / distance);

  // Clamp the zoom level to the maximum zoom level
  zoom = Math.min(zoom, ZOOM_MAX);
  // Return the zoom level
  return Math.floor(zoom);
}

// You need a function to calculate the distance between two points in kilometers
export function getDistance(coord1: number[], coord2: number[]): number {
  const R = 6371; // Radius of the earth in km
  const lat1 = coord1[0] * (Math.PI / 180);
  const lat2 = coord2[0] * (Math.PI / 180);
  const latDiff = (coord2[0] - coord1[0]) * (Math.PI / 180);
  const lonDiff = (coord2[1] - coord1[1]) * (Math.PI / 180);

  const a = Math.sin(latDiff / 2) * Math.sin(latDiff / 2) +
    Math.cos(lat1) * Math.cos(lat2) *
    Math.sin(lonDiff / 2) * Math.sin(lonDiff / 2);
  const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
  // Distance in km
  return R * c;
}

export function calculateMidpoint(source: number[], destination: number[]): [number, number] {
  const [lat1, lon1] = source.map((coord) => (coord * Math.PI) / 180); // convert degrees to radians
  const [lat2, lon2] = destination.map((coord) => (coord * Math.PI) / 180);

  const dLon = lon2 - lon1;

  const Bx = Math.cos(lat2) * Math.cos(dLon);
  const By = Math.cos(lat2) * Math.sin(dLon);

  const midLat = Math.atan2(Math.sin(lat1) + Math.sin(lat2), Math.sqrt((Math.cos(lat1) + Bx) * (Math.cos(lat1) + Bx) + By * By));
  const midLon = lon1 + Math.atan2(By, Math.cos(lat1) + Bx);

  return [midLat * (180 / Math.PI), midLon * (180 / Math.PI)]; // convert radians to degrees
}
