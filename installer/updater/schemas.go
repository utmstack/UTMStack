package updater

type Auth struct {
	ID  string `json:"id"`
	Key string `json:"key"`
}

type Release struct {
	ID        string `json:"id,omitempty"`
	Version   string `json:"version"`
	Changelog string `json:"changelog"`
	Images    string `json:"images"`
	Edition   string `json:"edition"`
}
