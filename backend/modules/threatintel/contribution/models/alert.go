package models

type Alert struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	DataType string   `json:"dataType"`
	Events   []*Event `json:"events"`
}
