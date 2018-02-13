package storage

import "github.com/go-xorm/xorm"

var (
	engine *xorm.Engine
)

func HostInsert(host *Host) error {
	exists, err := engine.Get(host)
	if err != nil {
		return err
	} else if exists {
		return ErrHostAlreadyExists(host.Name)
	}

	_, err = engine.Insert(host)
	return err
}

func HostDelete(host *Host) {
	// purge status info
	// purge fencer info
	// remove host record
	if host.Id > 0 {
		_, err := engine.Id(host.Id).Delete(host)
	} else {
		_, err := engine.Where("name=?", host.Name).Delete(host)
	}
}

func HostGet(host *Host) error {
	_, err := engine.Id(hostId).Get(&host)
	return err
}

func HostUpdate(host *Host) error {
	if err := HostGet(host); err != nil {
		return err
	}
	_, err := engine.Id(host.Id).Update(host)
	return err
}

func StateInsert() {
}

func StateDelete() {

}

func StateGet() {
}

func StateUpdate() {
}

func FencerInsert() {
}

func FencerDelete() {

}

func FencerGet() {
}

func FencerUpdate() {
}
