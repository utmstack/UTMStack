package updater

type Auth struct {
	ID  string `json:"id"`
	Key string `json:"key"`
}

type MasterVersion struct {
	ID                string             `json:"id"`
	VersionName       string             `json:"version_name"`
	Changelog         string             `json:"changelog"`
	ComponentVersions []ComponentVersion `json:"component_versions"`
}

type ComponentVersion struct {
	ID          string    `json:"id"`
	VersionName string    `json:"version_name"`
	Changelog   string    `json:"changelog,omitempty"`
	Edition     string    `json:"edition"`
	Component   Component `json:"component"`
	Scripts     []Script  `json:"scripts,omitempty"`
	Files       []File    `json:"files,omitempty"`
}

type Component struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Script struct {
	ID     string `json:"id"`
	Script string `json:"script"`
}

type File struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Hash            string `json:"hash"`
	Url             string `json:"-"`
	Binary          bool   `json:"binary"`
	DestinationPath string `json:"destination_path"`
	CpuArch         string `json:"cpu_arch"`
	Os              string `json:"os"`
	ReplacePrevious bool   `json:"replace_previous"`
}

func GetVersionsFromMaster(master MasterVersion) map[string]string {
	versions := make(map[string]string)
	versions["master"] = master.VersionName

	for _, v := range master.ComponentVersions {
		versions[v.Component.Name] = v.VersionName
	}

	return versions
}
