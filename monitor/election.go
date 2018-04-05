package monitor

import (
	"context"
	"errors"
	"time"

	"github.com/go-xorm/xorm"
)

const (
	defaultElectionName      = "themisLeader"
	defaultTermOfCampaign    = 30 // seconds
	defaultKeepClaimInterval = defaultTermOfCampaign / 5
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
		}

		if !succ {
			// wait and campaign again
			plog.Debugf("%s campaign failed, sleep %d seconds and try again.",
				e.LeaderName, defaultTermOfCampaign)
			time.Sleep(defaultTermOfCampaign * time.Second)
			continue
		} else {
			// start a goroutine to keep claim every 5 seconds.
			plog.Infof("%s campaign successed.", e.LeaderName)
			go keepClaim(ctx, e.Engine, e.LeaderName, quit)
			break
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
		case <-time.After(defaultKeepClaimInterval * time.Second):
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

// Quit end the term of leader
func (e *Election) Quit() error {
	isLeader, err := e.IsLeader()
	if err != nil {
		plog.Info("Failed to judge if we are the leader: ", err)
	}

	if isLeader {
		sql := `DELETE FROM election_record where election_name=? and leader_name=?`
		_, err = e.Engine.Exec(sql, defaultElectionName, e.LeaderName)
	}
	return err
}

func campaign(engine *xorm.Engine, leader string) (bool, error) {
	sql := `INSERT IGNORE INTO election_record (election_name, leader_name, last_update) VALUES (?, ?, ?)
			ON DUPLICATE KEY UPDATE
			leader_name = IF(last_update < DATE_SUB(VALUES(last_update), INTERVAL ? SECOND), VALUES(leader_name), leader_name),
			last_update = IF(leader_name = VALUES(leader_name), VALUES(last_update), last_update)`
	res, err := engine.Exec(sql, defaultElectionName, leader, time.Now(), defaultTermOfCampaign)

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
