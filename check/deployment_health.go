package check

import (
	"fmt"
	"encoding/json"
	flag "github.com/spf13/pflag"
	"k8s.io/kubernetes/pkg/apis/extensions"
)

type DeploymentHealth struct {
	BaseCheck
	warnLevel   *float32
	critLevel   *float32
	deployment  *extensions.Deployment
	commandLine *flag.FlagSet
}

//NewDeploymentHealth creates a new deployment health check
func NewDeploymentHealth(config CheckConfig) (Check, error) {
	dh := DeploymentHealth{}
	dh.Config = config

	// process flags
	commandLine := flag.NewFlagSet(config.Id, flag.ContinueOnError)
	dh.warnLevel = commandLine.Float32P("warn", "w", 0.9, "Percent of healthy pods to warn at")
	dh.critLevel = commandLine.Float32P("crit", "c", 0.8, "Percent of healthy pods to alert critical at")
	err := commandLine.Parse(config.Argv[1:])
	dh.commandLine = commandLine
	if err != nil {
		return &dh, err
	}
	if *dh.warnLevel <= float32(0) || *dh.warnLevel > float32(1) {
		return &dh, fmt.Errorf("--warn must be > 0 and <= 1")
	}
	if *dh.critLevel <= float32(0) || *dh.critLevel > float32(1) {
		return &dh, fmt.Errorf("--cirt must be > 0 and <= 1")
	}

	return &dh, nil
}

func (dh *DeploymentHealth) Usage() CheckUsage {
	return CheckUsage{
		Description: `Checks deployment pod levels via status info provided by Kubernetes. Provides full deployment status object in result output`,
		Flags: dh.commandLine.FlagUsages(),
	}
}

func (dh *DeploymentHealth) Update(resource interface{}) {
	dh.deployment = resource.(*extensions.Deployment)
}

func (dh *DeploymentHealth) Execute() (CheckResult, error) {
	res := NewCheckResultFromConfig(dh.Config)

	status := dh.deployment.Status

	res.Status = OK
	level := float32(status.AvailableReplicas) / float32(status.Replicas)
	if level <= *dh.critLevel {
		res.Status = CRITICAL
	} else if level <= *dh.warnLevel {
		res.Status = WARN
	}

	o, _ := json.MarshalIndent(status, "", "  ")
	res.Output = string(o)

	return res, nil
}

// register factory
func init() {
	RegisterCheck("deployment_health", NewDeploymentHealth, []string{"deployment"})
}
