package storage

import "time"

func init() {
	allTables = append(allTables,
		new(ElectionRecord),
		new(Host),
		new(HostState),
		new(HostFencer),
	)
}

type ElectionRecord struct {
	Id           uint32    `xorm:"INT notnull autoincr pk"`
	ElectionName string    `xorm:"VARCHAR(32) notnull default '' unique"`
	LeaderName   string    `xorm:"VARCHAR(32) notnull default ''"`
	LastUpdate   time.Time `xorm:"TIMESTAMP notnull default CURRENT_TIMESTAMP"`
}

type Host struct {
	Id     int    `json:"id" xorm:"pk autoincr"`
	Name   string `json:"name" binding:"required" xorm:"unique notnull"`
	Status string `json:"status"`
}

type HostState struct {
	Id          int    `json:"id" xorm:"pk autoincr"`
	HostId      int    `json:"host_id"`
	Tag         string `json:"tag" binding:"required" xorm:"varchar(64) notnull"`
	FailedTimes int    `json:"failed_times" xorm:"default 0"`
}

type HostFencer struct {
	Id       int    `json:"id" xorm:"pk autoincr"`
	HostId   int    `json:"host_id"`
	Type     string `json:"type" xorm:"VARCHAR(16) notnull"`
	Host     string `json:"host" binding:"required" xorm:"varchar(64) notnull"`
	Port     int    `json:"port" xorm:"default 623"`
	Username string `json:"username" binding:"required" xorm:"varchar(64) notnull"`
	Password string `json:"password" binding:"required" xorm:"varchar(64) notnull"`
}
