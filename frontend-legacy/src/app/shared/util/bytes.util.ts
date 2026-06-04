export function convertBytesToGB(bytes: number): number {
  const gigabytes = bytes / (1024 * 1024 * 1024);
  return parseFloat(gigabytes.toFixed(2));
}

