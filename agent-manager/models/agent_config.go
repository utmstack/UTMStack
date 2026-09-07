package models

// GlobalAgentID is the AgentConfig.AgentID value that holds a key's
// fleet-wide default, applied to every agent that has no more specific
// override for that key.
const GlobalAgentID uint = 0

// AgentConfigRevision is the single monotonic revision counter for a
// config key, shared across its fleet-wide default row and every
// per-agent override row in AgentConfig. Called "revision", not "version",
// to avoid reading as the agent's own software version (models.Agent.
// Version) — this counts edits to a config key, unrelated. Kept separate
// from AgentConfig instead of one counter per row so an agent's
// locally-remembered revision never looks like it went backwards when a
// per-agent override is removed and the agent falls back to the
// fleet-wide default: any write under this key, from any scope, bumps the
// same counter, so the agent is always told to re-fetch its effective
// content.
type AgentConfigRevision struct {
	Key      string `gorm:"primaryKey"`
	Revision uint64 `gorm:"not null;default:1"`
}

// AgentConfig holds the actual content for one key, scoped either to a
// single agent (AgentID = that agent's id) or to every agent that has no
// override for this key (AgentID = GlobalAgentID).
type AgentConfig struct {
	Key     string `gorm:"primaryKey"`
	AgentID uint   `gorm:"primaryKey;default:0"`
	Content string `gorm:"type:text"`
}
