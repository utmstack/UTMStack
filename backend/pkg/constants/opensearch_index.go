package constants

const (
	OS_INDEX_FIELD_LIMIT                = 50000
	OS_INDEX_NUMBER_OF_SHARDS           = 3
	OS_INDEX_NUMBER_OF_REPLICAS         = 0
	OS_TEMPLATE_PRIORITY_FIELD_LIMIT    = 1 // utmstack-field-limit (v11-log-*/v11-alert-*)
	OS_TEMPLATE_PRIORITY_CUSTOM_PATTERN = 2 // per-datasource custom index patterns (most specific)
)
