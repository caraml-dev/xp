package main

import (
	"github.com/caraml-dev/turing/engines/experiment/plugin/rpc"
	"github.com/caraml-dev/turing/engines/experiment/plugin/rpc/manager"
	"github.com/caraml-dev/turing/engines/experiment/plugin/rpc/runner"
	_manager "github.com/caraml-dev/xp/plugins/turing/manager"
	_runner "github.com/caraml-dev/xp/plugins/turing/runner"

	_ "github.com/caraml-dev/turing/engines/experiment/log/hclog"
)

func main() {
	rpc.Serve(&rpc.ClientServices{
		Manager: manager.NewConfigurableCustomExperimentManager(_manager.NewExperimentManager),
		Runner:  runner.NewConfigurableExperimentRunner(_runner.NewExperimentRunner),
	})
}
