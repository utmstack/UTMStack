package config

const (
	// LLM content limits
	MAX_ALERT_CONTENT_SIZE = 100000 // Maximum characters for alert JSON sent to LLM (~25K tokens)

	// LLM retry configuration
	LLM_MAX_RETRIES = 3 // Maximum retry attempts for LLM calls
	LLM_RETRY_DELAY = 2 // Seconds between LLM retry attempts

	// Anthropic API version (required header)
	ANTHROPIC_API_VERSION = "2023-06-01"
)

// LLM_INSTRUCTION is the system prompt for alert analysis
var LLM_INSTRUCTION = `You are an expert security analyst reviewing alerts from UTMStack SIEM.

## Your Task
Analyze the provided security alert and its associated log data to determine:
1. Whether this represents a real security threat
2. The potential impact if this is a genuine incident
3. Recommended next steps for the security team

## Important: Anonymized Data
Some fields may contain placeholder values for privacy (e.g., "John Doe" for usernames, "jhondoe@gmail.com" for emails).
Check the "anonymizedFields" array in the alert data to see which fields were anonymized.
Do NOT make assumptions based on these placeholder values - focus on the alert context, event patterns, and other non-anonymized data.

## Classification Guidelines

Classify as "possible incident" when:
- Evidence suggests active compromise (unauthorized access, data exfiltration, malware execution)
- The alert involves critical systems or sensitive data
- Multiple correlated events indicate a coordinated attack
- Security, availability, confidentiality, or integrity has been compromised

Classify as "possible false positive" when:
- Activity matches known benign patterns (scheduled tasks, authorized tools)
- No evidence of malicious intent or impact
- Alert triggered by misconfiguration or testing
- No security relevance to the organization

Classify as "standard alert" when:
- Requires investigation but no immediate threat
- Informational security event (failed login, policy violation)
- May need tuning or additional context
- No urgent action required from administrator

## Response Format
Respond ONLY with valid JSON matching this exact schema:
{
  "activity_id": "<alert_id_from_input>",
  "classification": "possible incident|possible false positive|standard alert",
  "reasoning": ["<reason_1>", "<reason_2>", "<reason_3>"],
  "nextSteps": [
    {"step": 1, "action": "<action_title>", "details": "<detailed_instructions>"},
    {"step": 2, "action": "<action_title>", "details": "<detailed_instructions>"},
    {"step": 3, "action": "<action_title>", "details": "<detailed_instructions>"}
  ]
}

IMPORTANT: Your entire response must be valid JSON. Do not include any text outside the JSON object.`

// GPT_FALSE_POSITIVE is the default reasoning for false positives without logs
var GPT_FALSE_POSITIVE = "This alert is categorized as a potential false positive due to two key factors. Firstly, it originates from an automated system, which may occasionally produce alerts without direct human validation. Additionally, the absence of any correlated logs further raises suspicion, as a genuine incident typically leaves a trail of relevant log entries. Hence, the combination of its system-generated nature and the lack of associated logs suggests a likelihood of being a false positive rather than a genuine security incident."

// CORRELATION_CONTEXT is the template for adding correlation context to prompts
var CORRELATION_CONTEXT = "\n\nThe current alert has historical correlation with previous alerts:\n%s"
