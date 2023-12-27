 export const environment = {
  production: true,
  SERVER_API_URL: ``,
  WEBSOCKET_URL: ``,
  SESSION_AUTH_TOKEN: window.location.host.split(':')[0].toLocaleUpperCase(),
  SERVER_API_CONTEXT: '/',
  BUILD_TIMESTAMP: new Date().getTime(),
  DEBUG_INFO_ENABLED: true,
  VERSION: '0.0.1'
};
