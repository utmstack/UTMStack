package config

const (
	// HTTP API Server
	HTTP_API_PORT = "8090"

	// API Endpoints
	API_ALERT_ENDPOINT                  = "/api/elasticsearch/search"
	API_ALERT_STATUS_ENDPOINT           = "/api/utm-alerts/status"
	API_INCIDENT_ENDPOINT               = "/api/utm-incidents"
	API_INCIDENT_ADD_NEW_ALERT_ENDPOINT = "/api/utm-incidents/add-alerts"
	API_ALERT_INFO_PARAMS               = "?page=0&size=25&top=10000&indexPattern="
	ELASTIC_DOC_ENDPOINT                = "/_doc/"
	ELASTIC_UPDATE_BY_QUERY_ENDPOINT    = "/_update_by_query"

	// Index patterns
	ALERT_INDEX_PATTERN = "v11-alert-*"
	LOGS_INDEX_PATTERN  = "v11-log-*"
	SOC_AI_INDEX        = "v11-soc-ai"

	// Status codes
	API_ALERT_COMPLETED_STATUS_CODE = 5

	// Timeouts (in seconds)
	HTTP_GPT_TIMEOUT = 90 // Timeout for LLM API calls
	HTTP_TIMEOUT     = 30 // Timeout for general HTTP calls

	// Startup timing
	CONFIG_STARTUP_DELAY = 2 // Seconds to wait for config system to initialize
	ERROR_EXIT_DELAY     = 5 // Seconds to wait before exiting on fatal error

	// Config polling
	TIME_FOR_GET_CONFIG = 10 // Seconds between config refresh attempts
	CLEANER_DELAY       = 10

	// Correlation limits
	CORRELATION_MAX_ALERTS = 100 // Maximum historical alerts to fetch for correlation

	// Separators
	LOGS_SEPARATOR = "\n\n" // Double newline - cleaner for display
)
