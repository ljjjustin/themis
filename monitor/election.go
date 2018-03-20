package monitor

import (
	"context"
	"time"

	_ "github.com/go-sql-driver/mysql"
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
func (e *Election) Campaign(ctx context.Context) error {

	for {
		succ, err := campaign(e.Engine, e.LeaderName)
		if err != nil {
			return err
		}

		if succ {
			// update record every 5 seconds if we became leader.
			go func(ctx context.Context, engine *xorm.Engine, leader string) error {
				for {
					_, err := campaign(engine, leader)
					if err != nil {
						return err
					}
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(5 * time.Second):
						continue
					}
				}
			}(ctx, e.Engine, e.LeaderName)
			return nil
		} else {
			// wait 20 seconds and campaign again
			time.Sleep(20 * time.Second)
		}
	}
}

// IsLeader query engine if we are the Leader.
func (e *Election) IsLeader(ctx context.Context) (bool, error) {

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
