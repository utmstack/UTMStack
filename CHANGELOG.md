# UTMStack 10.7.3 Release Notes
-- Implemented backend support for filtering compliance reports based on active integrations, optimizing query performance and data retrieval.
-- Introduced new compliance reports aligned with the PCI DSS standard to expand auditing capabilities.
-- Added support for creating and updating tag-based rules with dynamic conditions.

### Bug Fixes
-- Improved exception handling in `automaticReview` to prevent the process from stopping due to errors, ensuring the system continues evaluating alerts even if a specific rule fails.
-- Improved operator selection for more accurate and consistent filtering.