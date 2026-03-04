package models

type DataRetention struct {
	Retention int `json:"retention"`
}

// MSGDS represents a message with its data source, used by collectors.
type MSGDS struct {
	DataSource string
	Message    string
}
