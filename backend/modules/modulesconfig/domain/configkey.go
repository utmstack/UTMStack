package domain

type ModuleConfigurationKey struct {
	GroupID         int64
	ConfKey         string
	ConfName        string
	ConfValue       string
	ConfDescription string
	ConfDataType    string
	ConfOptions     string
	ConfVisibility  string
	ConfRequired    bool
}

func (k ModuleConfigurationKey) ToConfiguration(groupID int64) UtmModuleGroupConfiguration {
	return UtmModuleGroupConfiguration{
		GroupID:         groupID,
		ConfKey:         k.ConfKey,
		ConfName:        k.ConfName,
		ConfValue:       k.ConfValue,
		ConfDescription: k.ConfDescription,
		ConfDataType:    k.ConfDataType,
		ConfRequired:    k.ConfRequired,
		ConfOptions:     k.ConfOptions,
		ConfVisibility:  k.ConfVisibility,
	}
}
