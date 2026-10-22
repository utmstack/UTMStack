package domain

import (
	"time"
)

// AnalystConversation is one AI-analyst chat, scoped to its tenant and creator.
// The frontend generates its own base36 id before the first save, so Id is a
// string rather than a UUID — the backend never rewrites it.
//
// Turns and History are full JSON snapshots of the conversation state. A turn
// and the history it produced are written together at the end of the agent
// loop, so a mid-stream crash loses nothing — the conversation is simply
// shorter.
type AnalystConversation struct {
	ID           string    `gorm:"column:id;size:64;primaryKey" json:"id"`
	TenantID     string    `gorm:"column:tenant_id;size:100;not null;index;uniqueIndex:idx_ti_analyst_conv_tenant_id,priority:1" json:"-"`
	UserID       string    `gorm:"column:user_id;size:100;not null;index" json:"userId"`
	Title        string    `gorm:"column:title;size:255;not null" json:"title"`
	Turns        string    `gorm:"column:turns;type:text;not null;default:'[]'" json:"turns"`
	History      string    `gorm:"column:history;type:text;not null;default:'[]'" json:"history"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;index" json:"updatedAt"`
	CreatedDate  time.Time `gorm:"column:created_date" json:"createdDate"`
	ModifiedDate time.Time `gorm:"column:modified_date" json:"modifiedDate"`
}

func (AnalystConversation) TableName() string { return "ti_analyst_conversation" }
