package database

import (
	"fmt"
	"time"

	"github.com/coreos/pkg/capnslog"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"

	"github.com/ljjjustin/themis/config"
)

var plog = capnslog.NewPackageLogger("github.com/ljjjustin/themis", "database")

var (
	engine    *xorm.Engine
	allTables []interface{}
)

func Engine(cfg *config.DatabaseConfig) *xorm.Engine {
	var err error

	if engine == nil {
		url := ""
		switch cfg.Driver {
		case "mysql":
			url = fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true",
				cfg.Username, cfg.Password, cfg.Host, cfg.Name)
		case "sqlite3":
			url = cfg.Path
		default:
			plog.Fatal("unsupported database driver, check your configurations")
		}
		engine, err = xorm.NewEngine(cfg.Driver, url)
		if err != nil {
			plog.Fatal(err)
		}
		engine.DatabaseTZ = time.Local
		engine.TZLocation = time.Local
		// fast fail if if we can not connect to database
		err = engine.Ping()
		if err != nil {
			plog.Fatal(err)
		}
		// register and update database models
		err := engine.Sync2(allTables...)
		if err != nil {
			plog.Fatal(err)
		}
	}
	return engine
}

func HostInsert(host *Host) error {
	_, err := engine.Insert(host)
	return err
}

func HostGetAll() ([]*Host, error) {
	hosts := make([]*Host, 0)

	err := engine.Iterate(new(Host),
		func(i int, bean interface{}) error {
			host := bean.(*Host)
			hosts = append(hosts, host)
			return nil
		})
	return hosts, err
}

func HostGetById(id int) (*Host, error) {
	var host = Host{Id: id}

	exist, err := engine.Get(&host)
	if err != nil {
		return nil, err
	} else if exist {
		return &host, nil
	} else {
		return nil, nil
	}
}

func HostGetByName(hostname string) (*Host, error) {
	var host = Host{Name: hostname}

	exist, err := engine.Get(&host)
	if err != nil {
		return nil, err
	} else if exist {
		return &host, nil
	} else {
		return nil, nil
	}
}

func HostUpdate(id int, host *Host) error {
	_, err := engine.ID(id).Update(host)
	return err
}

func HostUpdateFields(host *Host, fields ...string) error {
	_, err := engine.ID(host.Id).Cols(fields...).Update(host)
	return err
}

func HostDelete(id int) error {
	_, err := engine.ID(id).Delete(new(Host))
	return err
}

func StateGetAll(hostId int) ([]*HostState, error) {
	states := make([]*HostState, 0)

	err := engine.Where("host_id=?", hostId).Iterate(new(HostState),
		func(i int, bean interface{}) error {
			state := bean.(*HostState)
			states = append(states, state)
			return nil
		})
	return states, err
}

func StateGetById(id int) (*HostState, error) {
	var state = HostState{Id: id}

	exist, err := engine.Get(&state)
	if err != nil {
		return nil, err
	} else if exist {
		return &state, nil
	} else {
		return nil, nil
	}
}

func StateInsert(state *HostState) error {
	_, err := engine.Insert(state)
	return err
}

func StateUpdate(id int, state *HostState) error {
	_, err := engine.ID(id).Update(state)
	return err
}

func StateUpdateFields(state *HostState, fields ...string) error {
	_, err := engine.ID(state.Id).Cols(fields...).Update(state)
	return err
}

func StateDelete(id int) error {
	_, err := engine.ID(id).Delete(new(HostState))
	return err
}

func FencerGetAll() ([]*HostFencer, error) {
	fencers := make([]*HostFencer, 0)

	err := engine.Iterate(new(HostFencer),
		func(i int, bean interface{}) error {
			fencer := bean.(*HostFencer)
			fencers = append(fencers, fencer)
			return nil
		})
	return fencers, err
}

func FencerGetByHost(hostId int) ([]*HostFencer, error) {
	fencers := make([]*HostFencer, 0)

	err := engine.Where("host_id=?", hostId).Iterate(new(HostFencer),
		func(i int, bean interface{}) error {
			fencer := bean.(*HostFencer)
			fencers = append(fencers, fencer)
			return nil
		})
	return fencers, err
}

func FencerGetById(id int) (*HostFencer, error) {
	var fencer = HostFencer{Id: id}

	exist, err := engine.Get(&fencer)
	if err != nil {
		return nil, err
	} else if exist {
		return &fencer, nil
	} else {
		return nil, nil
	}
}

func FencerInsert(fencer *HostFencer) error {
	_, err := engine.Insert(fencer)
	return err
}

func FencerUpdate(id int, fencer *HostFencer) error {
	_, err := engine.ID(id).Update(fencer)
	return err
}

func FencerDelete(id int) error {
	_, err := engine.ID(id).Delete(new(HostFencer))
	return err
}
