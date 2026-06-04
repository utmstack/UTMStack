// These constants are injected via webpack environment variables.
// You can add more variables in webpack.common.js or in profile specific webpack.<dev|prod>.js files.
// If you change the values in the webpack config files, you need to re run webpack to update the application

import {environment} from '../environments/environment';

export const VERSION = environment.VERSION;
export const SESSION_AUTH_TOKEN = environment.SESSION_AUTH_TOKEN + '_AUTH_TOKEN';
export const COOKIE_AUTH_TOKEN = 'utmauth';
export const SESSION_PRE_AUTH_TOKEN = environment.SESSION_AUTH_TOKEN + '_PRE_AUTH_TOKEN';
export const DEBUG_INFO_ENABLED: boolean = !!environment.DEBUG_INFO_ENABLED;
export const SERVER_API_URL = environment.SERVER_API_URL + environment.SERVER_API_CONTEXT;
export const SERVER_API_CONTEXT = environment.SERVER_API_CONTEXT;
export const WS_SERVER_API_URL = environment.WEBSOCKET_URL;
export const BUILD_TIMESTAMP = environment.BUILD_TIMESTAMP;

export const ACCESS_KEY =  'Utm-Internal-Key';
