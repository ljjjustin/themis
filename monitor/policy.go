package monitor

import "github.com/coreos/pkg/capnslog"

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

var plog = capnslog.NewPackageLogger("github.com/ljjjustin/themis", "policy")
