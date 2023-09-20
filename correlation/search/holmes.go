package search

import (
	"github.com/AtlasInsideCorp/UTMStackCorrelationEngine/utils"
	"github.com/quantfall/holmes"
)

var h = holmes.New(utils.GetConfig().ErrorLevel, "SEARCH")
