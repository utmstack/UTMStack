package search

import (
	"github.com/utmstack/UTMStack/correlation/utils"
	"github.com/quantfall/holmes"
)

var h = holmes.New(utils.GetConfig().ErrorLevel, "SEARCH")
