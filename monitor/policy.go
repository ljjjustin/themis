package monitor

import (
	"github.com/ljjjustin/themis/database"
)

const (
	flagManagement uint = 1 << 2
	flagStorage    uint = 1 << 1
	flagNetwork    uint = 1 << 0
)

var (
	flagTagMap = map[string]uint{
		"management": flagManagement,
		"storage":    flagStorage,
		"network":    flagNetwork,
	}
	decisionMatrix = []bool{
		/* +------------+----------+-----------+--------------+ */
		/* | Management | Storage  | Network   |    Fence     | */
		/* +------------+----------+-----------+--------------+ */
		/* | good       | good     | good      | */ false, /* | */
		/* | good       | good     | bad       | */ true, /*  | */
		/* | good       | bad      | good      | */ true, /*  | */
		/* | good       | bad      | bad       | */ true, /*  | */
		/* | bad        | good     | good      | */ false, /* | */
		/* | bad        | good     | bad       | */ true, /*  | */
		/* | bad        | bad      | good      | */ true, /*  | */
		/* | bad        | bad      | bad       | */ true, /*  | */
		/* +-----------------------------------+--------------+ */
	}
)

type PolicyEngine struct {
}

func NewPolicyEngine() *PolicyEngine {
	return &PolicyEngine{}
}

func isDown(state *database.HostState) bool {
	// FIXME: judge if one network is down.
	return true
}

func getDecision(states []*database.HostState) bool {
	var decision uint = 0
	for _, s := range states {
		if isDown(s) {
			decision |= flagTagMap[s.Tag]
		}
	}
	return decisionMatrix[decision]
}
