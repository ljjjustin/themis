package monitor

import (
	"context"
	"time"

	"github.com/go-xorm/xorm"
)

const (
	defaultElectionName   = "themisLeader"
	defaultTermOfCampaign = 30 // seconds
)

type Election struct {
	LeaderName string
	Engine     *xorm.Engine
}

func NewElection(name string, engine *xorm.Engine) *Election {
	return &Election{LeaderName: name, Engine: engine}
}

// Campaign puts a value as eligible for the election.
// It blocks until it is elected or an error occurs, or the context is cancelled.
func (e *Election) Campaign(ctx context.Context) <-chan error {
	quit := make(chan error, 1)

	for {
		succ, err := e.Proclaim()
		if err != nil {
			plog.Info("Campaign failed: ", err)
			quit <- err
			break
		} else if succ {
			plog.Infof("%s campaign successed.", e.LeaderName)
			break
		}
		plog.Debugf("%s campaign failed, sleep %d seconds and try again.",
			e.LeaderName, defaultTermOfCampaign)
		// wait and campaign again
		select {
		case <-ctx.Done():
			plog.Info("Quit campaign: ", ctx.Err())
			return quit
		case <-time.After(defaultTermOfCampaign * time.Second):
			continue
		}
	}

	return quit
}

func (e *Election) Proclaim() (bool, error) {
	sql := `INSERT IGNORE INTO election_record (election_name, leader_name, last_update) VALUES (?, ?, ?)
			ON DUPLICATE KEY UPDATE
			leader_name = IF(last_update < DATE_SUB(VALUES(last_update), INTERVAL ? SECOND), VALUES(leader_name), leader_name),
			last_update = IF(leader_name = VALUES(leader_name), VALUES(last_update), last_update)`
	res, err := e.Engine.Exec(sql, defaultElectionName, e.LeaderName, time.Now(), defaultTermOfCampaign)

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

// isLeader query engine if we are the Leader.
func (e *Election) isLeader() (bool, error) {

	sql := `SELECT COUNT(*) FROM election_record where election_name=? and leader_name=?`

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
	plog.Debug("quit election so that other node can become leader more quickly.")

	sql := `DELETE FROM election_record where election_name=? and leader_name=?`
	_, err := e.Engine.Exec(sql, defaultElectionName, e.LeaderName)

	return err
}
