# Saved Log Explorer searches

A saved search preserves the dataset, data type, time range, filters, search
text, and selected table columns in their current order. Columns are included
in the existing JSON state stored in the query's `filters` field; no backend
schema change is required.

Loading a saved search restores its columns. Older saved searches without a
`columns` array, or searches saved with an empty array, use automatic columns.
Non-string column values are ignored when restoring a search.

## Verification

From `frontend`, run:

```sh
npm test -- src/features/log-explorer/components/saved-search-columns.test.tsx
```

The integration tests exercise saving through the saved-query service and
loading into a freshly mounted Log Explorer, including column order, edits,
and compatibility with older saved searches. In a deployed UI, select custom
columns, save the search, and load it in a new tab to verify the same layout.
