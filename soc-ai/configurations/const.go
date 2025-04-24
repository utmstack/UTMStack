package configurations

import (
	"path/filepath"

	"github.com/utmstack/soc-ai/utils"
)

const (
	SOC_AI_SERVER_PORT                  = "8080"
	SOC_AI_SERVER_ENDPOINT              = "/process"
	API_ALERT_ENDPOINT                  = "/api/elasticsearch/search"
	API_ALERT_STATUS_ENDPOINT           = "/api/utm-alerts/status"
	API_INCIDENT_ENDPOINT               = "/api/utm-incidents"
	API_INCIDENT_ADD_NEW_ALERT_ENDPOINT = "/api/utm-incidents/add-alerts"
	API_ALERT_COMPLETED_STATUS_CODE     = 5
	API_ALERT_INFO_PARAMS               = "?page=0&size=25&top=10000&indexPattern="
	ELASTIC_DOC_ENDPOINT                = "/_doc/"
	ELASTIC_UPDATE_BY_QUERY_ENDPOINT    = "/_update_by_query"
	ALERT_INDEX_PATTERN                 = "alert-*"
	LOGS_INDEX_PATTERN                  = "log-*"
	SOC_AI_INDEX                        = "soc-ai"
	GPT_API_ENDPOINT                    = "https://api.openai.com/v1/chat/completions"
	TIME_FOR_GET_CONFIG                 = 10
	CLEANER_DELAY                       = 10
	MAX_ATTEMPS_TO_GPT                  = 3
	GPT_RESPONSE_TOKENS                 = 10
	HTTP_GPT_TIMEOUT                    = 90
	HTTP_TIMEOUT                        = 30
	LOGS_SEPARATOR                      = "[utm-logs-separator]"
)

var (
	AllowedGPTModels = map[string]int{
		"gpt-4.1":                1047576,
		"gpt-4.1-mini":           1047576,
		"gpt-4.1-nano":           1047576,
		"gpt-4.5-preview":        200000,
		"gpt-4o":                 128000,
		"gpt-4o-mini":            128000,
		"o1-preview":             200000,
		"o3":                     200000,
		"o4-mini":                200000,
		"gpt-3.5-turbo":          16385,
		"gpt-3.5-turbo-instruct": 4096,
		"gpt-3.5-turbo-16k":      16385,
		"gpt-3.5-turbo-0125":     4096,
		"gpt-3.5-turbo-1106":     16385,
	}
)

type SensitivePattern struct {
	Regexp    string
	FakeValue string
}

var (
	SensitivePatterns = map[string]SensitivePattern{
		"email": {Regexp: `([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`, FakeValue: "jhondoe@gmail.com"},
		//"ipv4":  `(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)`,
	}
	GPT_INSTRUCTION    = "You are an expert security engineer. Perform a deep analysis of an alert created by a SIEM and the logs related to it. Determine if the alert could be an actual potential threat or not and explain why. Provide a description that shows a deep understanding of the alert based on a deep analysis of its logs and estimate the risk to the systems affected. Classify the alert in the following manner: if the alert information is sufficient to determine that the security, availability, confidentiality, or integrity of the systems has being compromised, then classify it as \"possible incident\". If the alert does not pose a security risk to the organization or has no security relevance, classify it as \"possible false positive\". If the alert does not pose an imminent risk to the systems, requires no urgent action from an administrator, or requires not urgent review by an administrator, it should be classified as a \"standard alert\". You will also provide context-specific instructions for remediation, mitigation, or further investigation, related to the alert and logs analyzed. Your answer should be provided using the following JSON format and the total number of characters in your answer must not exceed 1500 words. Your entire answer must be inside this json format. {\"activity_id\":\"<activity_id>\",\"classification\":\"<classification>\",\"reasoning\":[\"<deep_reasoning>\"],\"nextSteps\":[{\"step\":1,\"action\":\"<action_1>\",\"details\":\"<action_1_details>\"},{\"step\":2,\"action\":\"<action_2>\",\"details\":\"<action_2_details>\"},{\"step\":3,\"action\":\"<action_3>\"]}Ensure that your entire answer adheres to the provided JSON format. The response should be valid JSON syntax and schema."
	GPT_FALSE_POSITIVE = "This alert is categorized as a potential false positive due to two key factors. Firstly, it originates from an automated system, which may occasionally produce alerts without direct human validation. Additionally, the absence of any correlated logs further raises suspicion, as a genuine incident typically leaves a trail of relevant log entries. Hence, the combination of its system-generated nature and the lack of associated logs suggests a likelihood of being a false positive rather than a genuine security incident."
)

func GetInternalKey() string {
	return utils.Getenv("INTERNAL_KEY", true)
}

func GetPanelServiceName() string {
	return utils.Getenv("PANEL_SERV_NAME", true)
}

func GetOpenSearchHost() string {
	return "http://" + utils.Getenv("OPENSEARCH_HOST", true)
}

func GetOpenSearchPort() string {
	return utils.Getenv("OPENSEARCH_PORT", true)
}

func GetAlertsDBPath() string {
	path, _ := utils.GetMyPath()
	return filepath.Join(path, "database", "alerts.sqlite3")
}
