package storage

import (
	"log"

	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
)

const (
	dbname = "apitest.db"
)

var (
	engine *xorm.Engine
)

func init() {
	var err error
	// create database connection
	engine, err = xorm.NewEngine("sqlite3", dbname)
	if err != nil {
		log.Fatal(err)
	}
	engine.ShowSQL(true)

	// register and update database models
	engine.Sync2(new(Host), new(HostState), new(HostFencer))
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

func HostUpdate(id int, host *Host) error {
	_, err := engine.ID(id).Update(host)
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

func StateDelete(id int) error {
	_, err := engine.ID(id).Delete(new(HostState))
	return err
}

func FencerGetAll(hostId int) ([]*HostFencer, error) {
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
