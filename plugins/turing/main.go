package main

import (
	_manager "github.com/caraml-dev/xp/plugins/turing/manager"
	_runner "github.com/caraml-dev/xp/plugins/turing/runner"
	"github.com/gojek/turing/engines/experiment/plugin/rpc"
	"github.com/gojek/turing/engines/experiment/plugin/rpc/manager"
	"github.com/gojek/turing/engines/experiment/plugin/rpc/runner"

	_ "github.com/gojek/turing/engines/experiment/log/hclog"
)

func main() {
	rpc.Serve(&rpc.ClientServices{
		Manager: manager.NewConfigurableCustomExperimentManager(_manager.NewExperimentManager),
		Runner:  runner.NewConfigurableExperimentRunner(_runner.NewExperimentRunner),
	})
}
