/**
 * Determines if an IPv4 address is private.
 * @param ip - The IP address as a string (e.g., "10.0.0.121").
 * @returns true if the IP is private, false otherwise.
 */
export function isPrivateIP(ip: string): boolean {
  if (!ip) {
    return false;
  }
  const parts = ip.split('.');
  if (parts.length !== 4) {
    // If it doesn't have 4 parts, it's not a valid IP address
    return false;
  }

  const octets = parts.map(part => parseInt(part, 10));

  if (octets.some(oct => isNaN(oct))) {
    return false;
  }

  // Private IP range 10.0.0.0/8: first octet is 10
  if (octets[0] === 10) { return true; }

  // Private IP range 172.16.0.0 - 172.31.255.255: first octet is 172 and second is between 16 and 31
  if (octets[0] === 172 && octets[1] >= 16 && octets[1] <= 31) { return true; }

  // Private IP range 192.168.0.0/16: first octet is 192 and second is 168
  if (octets[0] === 192 && octets[1] === 168) { return true; }

  // (Optional) Loopback range 127.0.0.0/8
  if (octets[0] === 127) { return true; }

  // If none of the above conditions are met, consider it a public IP address
  return false;
}
