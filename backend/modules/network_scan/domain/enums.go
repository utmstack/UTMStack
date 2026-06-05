package domain

type AssetStatus string

const (
	AssetStatusNew     AssetStatus = "NEW"
	AssetStatusCheck   AssetStatus = "CHECK"
	AssetStatusMissing AssetStatus = "MISSING"
)

type AssetRegisteredMode string

const (
	AssetRegisteredModeCustom     AssetRegisteredMode = "CUSTOM"
	AssetRegisteredModeDiscovered AssetRegisteredMode = "DISCOVERED"
	AssetRegisteredModeDynamic    AssetRegisteredMode = "DYNAMIC"
)

// UpdateLevel identifies who last updated an asset. Priority: AGENT > SCANNER > DATASOURCE.
type UpdateLevel string

const (
	UpdateLevelDataSource UpdateLevel = "DATASOURCE"
	UpdateLevelScanner    UpdateLevel = "SCANNER"
	UpdateLevelAgent      UpdateLevel = "AGENT"
)
