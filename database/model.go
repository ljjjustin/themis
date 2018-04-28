package database

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
	Id           uint32    `json:"id" xorm:"autoincr pk"`
	ElectionName string    `json:"election_name" xorm:"varchar(64) unique notnull"`
	LeaderName   string    `json:"leader_name" xorm:"varchar(64) notnull"`
	LastUpdate   time.Time `json:"last_update" xorm:"TIMESTAMP"`
}

type Host struct {
	Id        int       `json:"id" xorm:"pk autoincr"`
	Name      string    `json:"name" binding:"required" xorm:"varchar(64) unique notnull"`
	Status    string    `json:"status" xrom:"varchar(64) default 'initializing'"`
	Disabled  bool      `json:"disabled" xorm:"tinyint(1)" default false`
	UpdatedAt time.Time `json:"updated_at" xorm:"TIMESTAMP"`
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
	Type     string `json:"type" xorm:"varchar(32) notnull"`
	Host     string `json:"host" binding:"required" xorm:"varchar(64) notnull"`
	Port     int    `json:"port" xorm:"default 623"`
	Username string `json:"username" binding:"required" xorm:"varchar(64) notnull"`
	Password string `json:"password" binding:"required" xorm:"varchar(64) notnull"`
}
