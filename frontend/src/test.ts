// Karma entry: initializes the Angular testing environment and explicitly
// imports every spec (the Angular-7-era require.context auto-discovery no
// longer works under webpack 5 — see the import block below).

// zone.js 0.14 dropped the ./dist/zone-testing export path; the testing build
// is the bare 'zone.js/testing' entry (same reason polyfills.ts uses bare
// 'zone.js' since S17).
import 'zone.js/testing';
import { getTestBed } from '@angular/core/testing';
import {
  BrowserDynamicTestingModule,
  platformBrowserDynamicTesting
} from '@angular/platform-browser-dynamic/testing';

// First, initialize the Angular testing environment.
getTestBed().initTestEnvironment(
  BrowserDynamicTestingModule,
  platformBrowserDynamicTesting()
);

// S17 (webpack 5 / Angular 12+): the Angular-7-era `require.context('./',
// true, /\.spec\.ts$/)` no longer works - webpack 5's __webpack_require__ has
// no .context in the ESM-compiled test entry, so the loader crashed at runtime
// ("context is not a function"). List the specs explicitly (relative to this
// file in src/); add new *.spec.ts here when they are created.
import './app/data-management/file-management/shared/components/file-host-os/file-host-os.component.spec';
import './app/data-management/file-management/shared/components/file-object-type/file-object-type.component.spec';
import './app/incident-response/shared/component/action-builder/action-builder.component.spec';
import './app/incident-response/shared/component/action-conditional/action-conditional.component.spec';
import './app/incident-response/shared/component/action-terminal/action-terminal.component.spec';
import './app/incident-response/shared/component/agent-info/agent-info.component.spec';
import './app/shared/components/utm/util/utm-online-documentation/utm-online-documentation.component.spec';
