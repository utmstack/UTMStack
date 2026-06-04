package dto

import (
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
	"github.com/utmstack/utmstack/backend/pkg/eventprocessor"
)

// ToEventProcessorPayload serializes a module + its groups + the decrypted
// values into the wire shape event-processor-manager consumes. The legacy
// panel sent encrypted values straight through because event-processor shares
// the same ENCRYPTION_KEY; we follow that contract verbatim.
func ToEventProcessorPayload(m domain.UtmModule) eventprocessor.ModulePayload {
	payload := eventprocessor.ModulePayload{
		ID:                m.ID,
		ServerID:          m.ServerID,
		PrettyName:        m.PrettyName,
		ModuleName:        m.ModuleName,
		ModuleDescription: m.ModuleDescription,
		ModuleActive:      m.ModuleActive,
		ModuleIcon:        m.ModuleIcon,
		ModuleCategory:    m.ModuleCategory,
		LiteVersion:       m.LiteVersion,
		NeedsRestart:      m.NeedsRestart,
		IsActivatable:     m.IsActivatable,
		ModuleGroups:      make([]eventprocessor.ModuleGroup, 0, len(m.ModuleGroups)),
	}
	for _, g := range m.ModuleGroups {
		grp := eventprocessor.ModuleGroup{
			ID:                        g.ID,
			ModuleID:                  g.ModuleID,
			GroupName:                 g.GroupName,
			GroupDescription:          g.GroupDescription,
			Collector:                 g.Collector,
			ModuleGroupConfigurations: make([]eventprocessor.ModuleGroupConfig, 0, len(g.ModuleGroupConfigurations)),
		}
		for _, c := range g.ModuleGroupConfigurations {
			grp.ModuleGroupConfigurations = append(grp.ModuleGroupConfigurations, eventprocessor.ModuleGroupConfig{
				ID:              c.ID,
				GroupID:         c.GroupID,
				ConfKey:         c.ConfKey,
				ConfValue:       c.ConfValue,
				ConfName:        c.ConfName,
				ConfDescription: c.ConfDescription,
				ConfDataType:    c.ConfDataType,
				ConfRequired:    c.ConfRequired,
				ConfOptions:     c.ConfOptions,
				ConfVisibility:  c.ConfVisibility,
			})
		}
		payload.ModuleGroups = append(payload.ModuleGroups, grp)
	}
	return payload
}
