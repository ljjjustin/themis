package monitor

import (
	"context"
	"errors"
	"time"

	"github.com/go-xorm/xorm"
)

const (
	defaultElectionName = "themisLeader"
)

type Election struct {
	LeaderName string
	Engine     *xorm.Engine
}

func NewElection(name string, engine *xorm.Engine) *Election {
	return &Election{LeaderName: name, Engine: engine}
}

// Campaign puts a value as eligible for the election.
// It blocks until it is elected, an error occurs, or the context is cancelled.
func (e *Election) Campaign(ctx context.Context) <-chan error {
	quit := make(chan error, 1)

	for {
		succ, err := campaign(e.Engine, e.LeaderName)
		if err != nil {
			plog.Debug(err)
			quit <- err
			break
		} else if succ {
			// update authorization every 5 seconds.
			plog.Debugf("%s campaign successed.", e.LeaderName)
			go keepClaim(ctx, e.Engine, e.LeaderName, quit)
			break
		} else {
			// wait 20 seconds and campaign again
			plog.Debugf("%s campaign failed, sleep 20 seconds and try again.", e.LeaderName)
			time.Sleep(20 * time.Second)
			continue
		}
	}
	return quit
}

func keepClaim(ctx context.Context, engine *xorm.Engine, leader string, quit chan error) {
	for {
		select {
		case <-ctx.Done():
			plog.Info(ctx.Err())
			return
		case <-time.After(5 * time.Second):
		}
		plog.Debugf("%s updating authorization.", leader)
		succ, err := campaign(engine, leader)
		if err != nil {
			plog.Debug(err)
			quit <- err
			return
		} else if !succ {
			plog.Debugf("%s campaign failed, we are not leader now.", leader)
			quit <- errors.New("Leader changed.")
			return
		}
	}
}

// IsLeader query engine if we are the Leader.
func (e *Election) IsLeader() (bool, error) {

	sql := `SELECT COUNT(*) as is_leader FROM election_record where election_name=? and leader_name=?`

	res, err := e.Engine.Query(sql, defaultElectionName, e.LeaderName)

	if err != nil {
		return false, err
	}

	if len(res) > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func campaign(engine *xorm.Engine, leader string) (bool, error) {
	sql := `INSERT IGNORE INTO election_record (election_name, leader_name, last_update) VALUES (?, ?, ?)
			ON DUPLICATE KEY UPDATE
			leader_name = IF(last_update < DATE_SUB(VALUES(last_update), INTERVAL 30 SECOND), VALUES(leader_name), leader_name),
			last_update = IF(leader_name = VALUES(leader_name), VALUES(last_update), last_update)`
	res, err := engine.Exec(sql, defaultElectionName, leader, time.Now())

	if err != nil {
		return false, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if affected >= 1 {
		return true, nil
	} else {
		return false, nil
	}
}
