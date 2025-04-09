# UTMStack 10.7.2 Release Notes
-- Implemented backend support for filtering compliance reports based on active integrations, optimizing query performance and data retrieval.

### Bug Fixes
-- Improved exception handling in `automaticReview` to prevent the process from stopping due to errors, ensuring the system continues evaluating alerts even if a specific rule fails.
-- Improved operator selection for more accurate and consistent filtering.
-- Introduced new compliance reports aligned with the PCI DSS standard to expand auditing capabilities.
-- Enabled creation and update of tag-based rules with dynamic conditions.