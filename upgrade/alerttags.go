package upgrade

import (
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"gorm.io/gorm"
)

var (
	alertTagNamespace     = uuid.MustParse("402218d7-f216-56d0-a6bc-be7414c8aff9")
	alertTagRuleNamespace = uuid.MustParse("d65b0f32-73f4-5340-acf5-ae4df2dbaba0")
)

type v11AlertTag struct {
	ID          int64
	TagName     string
	TagColor    *string
	SystemOwner *bool
}

type v11AlertTagRule struct {
	ID          int64
	RuleName    string
	RuleDesc    string `gorm:"column:rule_description"`
	RuleCond    string `gorm:"column:rule_conditions"`
	AppliedTags string `gorm:"column:rule_applied_tags"`
	RuleActive  bool
	RuleDeleted bool
}

// convertAlertTags ports v11 tags. System tags are unified with the v12 seed
// row (they live in the nil tenant), so only customer tags are inserted here.
// Returns a v11 tag id -> v12 tag uuid map used to remap rule references.
func convertAlertTags(db *gorm.DB) (map[int64]string, error) {
	ids := make(map[int64]string)
	if !tableExists(db, "utm_alert_tag") {
		logSkipTable("utm_alert_tag")
		return ids, nil
	}

	var src []v11AlertTag
	if err := db.Raw(`SELECT id, tag_name, tag_color, system_owner FROM utm_alert_tag ORDER BY id`).Scan(&src).Error; err != nil {
		return ids, err
	}

	inserted := 0
	for _, t := range src {
		if isTrue(t.SystemOwner) {
			if id, ok := resolveSeedTag(db, t.TagName); ok {
				ids[t.ID] = id
				continue
			}
		}
		v12ID := uuid.NewSHA1(alertTagNamespace, []byte(strconv.FormatInt(t.ID, 10))).String()
		if err := db.Exec(
			`INSERT INTO alert_tag (id, tenant_id, tag_name, tag_color, system_owner)
			 VALUES (?, ?, ?, ?, ?)
			 ON CONFLICT (tenant_id, tag_name) DO UPDATE SET
				tag_color = EXCLUDED.tag_color,
				system_owner = EXCLUDED.system_owner`,
			v12ID, defaultTenantID, t.TagName, strFrom(t.TagColor), isTrue(t.SystemOwner),
		).Error; err != nil {
			return ids, err
		}
		ids[t.ID] = v12ID
		inserted++
	}

	catcher.Info("v11 upgrade: alert_tag converted", map[string]any{"inserted": inserted, "unified": len(src) - inserted})
	return ids, nil
}

// resolveSeedTag finds the v12 seed row for a system tag name (nil tenant).
func resolveSeedTag(db *gorm.DB, name string) (string, bool) {
	var id string
	err := db.Raw(`SELECT id FROM alert_tag WHERE tag_name = ? AND tenant_id = '00000000-0000-0000-0000-000000000000' LIMIT 1`, name).Scan(&id).Error
	return id, err == nil && id != ""
}

// convertAlertTagRules ports v11 rules and remaps rule_applied_tags from v11
// numeric ids to the v12 tag uuids; unknown ids are dropped.
func convertAlertTagRules(db *gorm.DB, tagIDs map[int64]string) (int, error) {
	if !tableExists(db, "utm_alert_tag_rule") {
		logSkipTable("utm_alert_tag_rule")
		return 0, nil
	}

	var src []v11AlertTagRule
	if err := db.Raw(`SELECT id, rule_name, rule_description, rule_conditions,
		rule_applied_tags, rule_active, rule_deleted FROM utm_alert_tag_rule ORDER BY id`).Scan(&src).Error; err != nil {
		return 0, err
	}

	count := 0
	for _, r := range src {
		applied := remapTagCSV(r.AppliedTags, tagIDs)
		if err := db.Exec(
			`INSERT INTO alert_tag_rule (id, tenant_id, rule_name, rule_description, rule_conditions, rule_applied_tags, rule_active, rule_deleted)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			 ON CONFLICT (tenant_id, rule_name) DO UPDATE SET
				rule_description = EXCLUDED.rule_description,
				rule_conditions = EXCLUDED.rule_conditions,
				rule_applied_tags = EXCLUDED.rule_applied_tags,
				rule_active = EXCLUDED.rule_active,
				rule_deleted = EXCLUDED.rule_deleted`,
			uuid.NewSHA1(alertTagRuleNamespace, []byte(strconv.FormatInt(r.ID, 10))).String(), defaultTenantID,
			r.RuleName, r.RuleDesc, r.RuleCond, applied, r.RuleActive, r.RuleDeleted,
		).Error; err != nil {
			return 0, err
		}
		count++
	}

	catcher.Info("v11 upgrade: alert_tag_rule converted", map[string]any{"rules": count})
	return count, nil
}

// remapTagCSV turns a v11 numeric tag id CSV into a v12 uuid CSV; unknown ids are dropped.
func remapTagCSV(csv string, tagIDs map[int64]string) string {
	parts := strings.Split(csv, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			continue
		}
		if v12, ok := tagIDs[id]; ok {
			out = append(out, v12)
		}
	}
	return strings.Join(out, ",")
}

func isTrue(b *bool) bool {
	return b != nil && *b
}
