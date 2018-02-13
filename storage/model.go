package storage

type Host struct {
	Id   uint32 `xorm:"INT pk autoincr"`
	Name string `xorm:"UNIQUE notnull"`
	UUID string `xorm:"uuid UNIQUE notnull"`
}

type HostState struct {
	Tag         string `xorm:"notnull"`
	Hostname    string `xorm:"notnull"`
	FailedTimes int32  `xorm:"default 0"`
}

type Fencer struct {
	Id       uint32 `xorm:"INT pk autoincr"`
	Type     string `xorm:"VARCHAR(16) notnull"`
	Host     string `xorm:"VARCHAR(64) notnull"`
	Port     int    `xorm:"INT notnull default 623"`
	Username string `xorm:"VARCHAR(32) notnull"`
	Password string `xorm:"VARCHAR(64) notnull"`
}
