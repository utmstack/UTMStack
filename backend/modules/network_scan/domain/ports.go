package domain

type UtmPorts struct {
	ID     uint64 `gorm:"primaryKey;autoIncrement;column:id"`
	ScanID uint64 `gorm:"column:scan_id;index"`
	Port   int    `gorm:"column:port"`
	TCP    string `gorm:"column:tcp;size:255"`
	UDP    string `gorm:"column:udp;size:255"`
}

func (UtmPorts) TableName() string { return "utm_ports" }
