/**
 * No-op stub for `intl-tel-input/build/js/utils`.
 *
 * ng2-tel-input (root app module) does a side-effect
 * `import 'intl-tel-input/build/js/utils'`. That file is a BARE IIFE that
 * registers `intlTelInputUtils` on top-level `this`:
 *     (function () { var aa = this; … m("intlTelInputUtils", data); })();
 * As a webpack 5.76 module (Angular 13) the bare IIFE's `this` is `undefined`
 * (strict scope), so it throws "Cannot use 'in' operator to search for
 * 'intlTelInputUtils' in undefined" at bootstrap. (webpack 5.50 / Angular 12
 * evaluated it as sloppy CJS where `this === window`, so it never threw.)
 *
 * Aliasing the import to this stub (webpack.custom.js) keeps webpack from
 * bundling/evaluating the throwing IIFE. The REAL utils.js is delivered
 * instead as a classic <script src="lib/intl-tel-input/utils.js"> tag in
 * src/index.html (file copied via the angular.json asset entry) — its intended
 * delivery mode, where top-level `this === window` and it sets
 * window.intlTelInputUtils for the intl-tel-input plugin.
 */
export {};
