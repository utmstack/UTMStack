//go:build windows

package ransomware

// t1490Rules — Windows recovery/backup-destruction commands (MITRE T1490).
// These are the rules the guard matched before the platform split; behavior
// is unchanged.
var t1490Rules = []t1490Rule{
	{"vssadmin_delete_shadows", []string{"vssadmin", "delete", "shadows"}},
	{"vssadmin_resize_shadowstorage", []string{"vssadmin", "resize", "shadowstorage"}},
	{"wmic_shadowcopy_delete", []string{"wmic", "shadowcopy", "delete"}},
	{"wbadmin_delete_catalog", []string{"wbadmin", "delete", "catalog"}},
	{"wbadmin_delete_backup", []string{"wbadmin", "delete", "backup"}},
	{"bcdedit_recovery_disable", []string{"bcdedit", "recoveryenabled", "no"}},
	{"bcdedit_ignore_failures", []string{"bcdedit", "bootstatuspolicy", "ignoreallfailures"}},
	{"reagentc_disable", []string{"reagentc", "/disable"}},
	{"diskshadow_delete", []string{"diskshadow", "delete"}},
}
