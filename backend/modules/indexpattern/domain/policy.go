package domain

const (
	PolicyID         = "utmstack_ism_policy"
	SnapshotRepoName = "utmstack_backups"
	SnapshotRepoPath = "/usr/share/opensearch/backups"
	SnapshotName     = "utmstack_snapshot"

	StateIngest     = "ingest"
	StateOpen       = "open"
	StateDelete     = "delete"
	StateBackup     = "backup"
	StateSafeDelete = "safe_delete"
)

type IndexPolicy struct {
	ID          string  `json:"_id,omitempty"`
	PrimaryTerm int64   `json:"_primary_term,omitempty"`
	SeqNo       int64   `json:"_seq_no,omitempty"`
	Policy      *Policy `json:"policy"`
}

type Policy struct {
	PolicyID     string        `json:"policy_id"`
	Description  string        `json:"description"`
	DefaultState string        `json:"default_state"`
	States       []State       `json:"states"`
	IsmTemplate  []IsmTemplate `json:"ism_template"`
}

type IsmTemplate struct {
	IndexPatterns []string `json:"index_patterns"`
	Priority      int      `json:"priority"`
}

type State struct {
	Name        string       `json:"name"`
	Actions     []Action     `json:"actions"`
	Transitions []Transition `json:"transitions"`
}

type Transition struct {
	StateName  string               `json:"state_name"`
	Conditions *TransitionCondition `json:"conditions,omitempty"`
}

type TransitionCondition struct {
	MinIndexAge string `json:"min_index_age"`
}

type Action struct {
	ForceMerge *ActionForceMerge `json:"force_merge,omitempty"`
	Delete     *ActionDelete     `json:"delete,omitempty"`
	Snapshot   *ActionSnapshot   `json:"snapshot,omitempty"`
	ReadWrite  *ActionReadWrite  `json:"read_write,omitempty"`
}

type ActionForceMerge struct {
	MaxNumSegments int `json:"max_num_segments"`
}

type ActionDelete struct{}

type ActionReadWrite struct{}

type ActionSnapshot struct {
	Repository string `json:"repository"`
	Snapshot   string `json:"snapshot"`
}

type AddPolicyRequest struct {
	PolicyID string `json:"policy_id"`
}

type UpdateManagedIndexPolicyConfiguration struct {
	PolicyID string         `json:"policy_id"`
	State    string         `json:"state"`
	Include  []IncludeState `json:"include"`
}

type IncludeState struct {
	State string `json:"state"`
}

type UpdateManagedIndexPolicyResponse struct {
	UpdatedIndices int           `json:"updated_indices"`
	Failures       bool          `json:"failures"`
	FailedIndices  []FailedIndex `json:"failed_indices"`
}

type FailedIndex struct {
	IndexName string `json:"index_name"`
	IndexUuid string `json:"index_uuid"`
	Reason    string `json:"reason"`
}

type RegisterRepositoryRequest struct {
	Type     string       `json:"type"`
	Settings RepoSettings `json:"settings"`
}

type RepoSettings struct {
	Location string `json:"location"`
}

type PolicySettings struct {
	SnapshotActive bool   `json:"snapshotActive"`
	DeleteAfter    string `json:"deleteAfter"`
}
