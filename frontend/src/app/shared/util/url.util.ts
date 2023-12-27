export function isSubdomainOfUtmstack(): boolean {
  const regex = /^(?!www\.)([a-z0-9-]+)\.utmstack\.com$/i;
  return regex.test(window.location.hostname);
}
