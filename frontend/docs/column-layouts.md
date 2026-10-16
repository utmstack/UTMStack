# Log Explorer column layouts

Each user tab keeps an independent selected-column order for each `(dataset, dataType)` pair. The all-data-types view has its own layout. Switching devices within a data type keeps the same layout. Preferences use existing user-specific browser tab storage and are not synchronized between browsers.

Existing selected columns migrate to the tab's previously selected scope. New scopes start with automatic columns drawn from populated result fields. Windows and FortiGate fields receive priority where applicable. Empty layouts mean automatic defaults.

**Restore default columns** in the fields sidebar resets only the current scope. The toast offers **Undo**, which restores the captured scope even after changing data type. Filters and the time range are independent of column layouts.

Combined with the saved-search fix, loading a saved search explicitly applies its columns to its destination scope.

Manual validation: switch between Windows and FortiGate, select/reorder columns independently, return to each type, reload, reset one type, switch types and undo. Confirm CSV order follows the active layout.
