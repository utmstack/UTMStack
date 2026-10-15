-- The v11.3.0 event-bag rename moved Windows event fields from the "log" bag
-- to the "event" bag. The Windows filter now emits event.eventCode, but the
-- seed in 005 set the source filter to log.eventCode. A match_phrase on
-- log.eventCode against new documents matches nothing, so Active Directory
-- never discovered users or events after the upgrade (issue #2894).
--
-- Point the filter at the new field. Guarded so re-running is a no-op once
-- the value has already been migrated.
UPDATE utm_source_filter
SET field = 'event.eventCode'
WHERE field = 'log.eventCode';
