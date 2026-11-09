// monaco-editor is loaded at RUNTIME via AMD from /assets/monaco/vs (see angular.json
// assets glob + ngx-monaco-editor's baseUrl). This file pulls the AMBIENT GLOBAL
// `monaco` namespace (monaco-editor's root monaco.d.ts, a pure global script) into the
// type program so `monaco.editor.*` / `monaco.languages.*` in code-editor.component.ts
// type-check. We intentionally do NOT `import * as monaco from 'monaco-editor'`: that
// would bundle the full ESM editor as a second runtime instance, and the SQL completion
// provider (registered via `monaco.languages.registerCompletionItemProvider`) would
// attach to the ESM instance while the DOM editor runs the AMD one -> provider lost.
/// <reference path="../node_modules/monaco-editor/monaco.d.ts" />
