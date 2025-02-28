package models

type Release struct {
	ID        string `json:"id,omitempty"`
	Version   string `json:"version"`
	Changelog string `json:"changelog"`
	Images    string `json:"images"`
	Edition   string `json:"edition"`
}

type Version struct {
	Version string `json:"version"`
}
