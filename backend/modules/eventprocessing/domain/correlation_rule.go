package domain

import (
	"fmt"
	"time"
)

type UtmCorrelationRules struct {
	ID int64 `gorm:"column:id;primaryKey"`

	RuleName      string `gorm:"column:rule_name;size:250;not null"`
	RuleAdversary string `gorm:"column:rule_adversary;size:25;not null"`

	// CIA impact scores — each 0-3.
	RuleConfidentiality int `gorm:"column:rule_confidentiality;not null"`
	RuleIntegrity       int `gorm:"column:rule_integrity;not null"`
	RuleAvailability    int `gorm:"column:rule_availability;not null"`

	RuleCategory    string `gorm:"column:rule_category;size:250;not null"`
	RuleTechnique   string `gorm:"column:rule_technique;size:500;not null"`
	RuleDescription string `gorm:"column:rule_description"`

	// rule_references_def: nullable TEXT — serialized JSON array.
	RuleReferencesDef string `gorm:"column:rule_references_def"`

	// rule_definition_def: NOT NULL TEXT — serialized rule definition JSON.
	RuleDefinitionDef string `gorm:"column:rule_definition_def;not null"`

	RuleLastUpdate *time.Time `gorm:"column:rule_last_update"`
	RuleActive     bool       `gorm:"column:rule_active;not null"`
	SystemOwner    bool       `gorm:"column:system_owner;not null"`

	// Nullable JSON TEXT columns for advanced rule configuration.
	//[deprecated] only keeped for compatibility
	AfterEventsDef   string `gorm:"column:rule_after_events_def"`
	//
	CorrelationDef   string `gorm:"column:rule_correlation_def"`
	RuleGroupByDef   string `gorm:"column:rule_group_by_def"`
	DeduplicateByDef string `gorm:"column:rule_deduplicate_by_def"`

	// Legacy-read-only: the rule bootstrap Preloads this association to recover
	// each rule's data-type names during the one-time DB->YAML migration. Never
	// written; the join table is dropped once migration completes.
	DataTypes []UtmDataTypes `gorm:"many2many:utm_group_rules_data_type;joinForeignKey:rule_id;joinReferences:data_type_id"`
}

func (self *UtmCorrelationRules) GetCorrelationDef() (string,error){
	var correlate string
	if self.AfterEventsDef!="" && self.CorrelationDef !=""{
		return "",fmt.Errorf("only one afterEvents or correlation allowed")
	}else if self.AfterEventsDef!="" {
		correlate = self.AfterEventsDef
	}else{
		correlate= self.CorrelationDef
	}
	return correlate,nil
}

func (UtmCorrelationRules) TableName() string { return "utm_correlation_rules" }
