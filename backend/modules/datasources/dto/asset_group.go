package dto

import (
	"time"

	"github.com/utmstack/utmstack/backend/modules/datasources/domain"
)

type AssetGroupDTO struct {
	ID               uint64           `json:"id,omitempty"`
	GroupName        string           `json:"groupName" binding:"required"`
	GroupDescription string           `json:"groupDescription,omitempty"`
	CreatedDate      *time.Time       `json:"createdDate,omitempty"`
	AssetsCount      int64            `json:"assetsCount"`
	Metrics          map[string]int64 `json:"metrics,omitempty"`
}

func ToAssetGroupDTO(g *domain.UtmAssetGroup, assetsCount int64) AssetGroupDTO {
	out := AssetGroupDTO{
		ID:               g.ID,
		GroupName:        g.GroupName,
		GroupDescription: g.GroupDescription,
		AssetsCount:      assetsCount,
	}
	if !g.CreatedDate.IsZero() {
		t := g.CreatedDate
		out.CreatedDate = &t
	}
	return out
}

func FromAssetGroupDTO(in AssetGroupDTO) *domain.UtmAssetGroup {
	g := &domain.UtmAssetGroup{
		ID:               in.ID,
		GroupName:        in.GroupName,
		GroupDescription: in.GroupDescription,
	}
	if in.CreatedDate != nil {
		g.CreatedDate = *in.CreatedDate
	}
	return g
}
