package check

import (
	"fmt"
	"encoding/json"
	flag "github.com/spf13/pflag"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type DeploymentStatus struct {
	BaseCheck
	warnLevel   *float32
	critLevel   *float32
	minReplicas *int32
	deployment  *v1beta1.Deployment
	commandLine *flag.FlagSet
}

//NewDeploymentStatus creates a new deployment health check
func NewDeploymentStatus(config CheckConfig) (Check, error) {
	dh := DeploymentStatus{}
	dh.Config = config

	// process flags
	commandLine := flag.NewFlagSet(config.Id, flag.ContinueOnError)
	dh.warnLevel = commandLine.Float32P("warn", "w", 0.9, "Percent of healthy (available) pods to warn at")
	dh.critLevel = commandLine.Float32P("crit", "c", 0.8, "Percent of healthy (available) pods to alert critical at")
	dh.minReplicas = commandLine.Int32P("min-configured-replicas", "m", 0, "Alert if a deployment gets configured with a replica count below X. Often users 'suspend' a service by setting 'replicas: 0'. Intended as a simple safeguard")
	err := commandLine.Parse(config.Argv[1:])
	dh.commandLine = commandLine
	if err != nil {
		return &dh, err
	}
	if *dh.warnLevel <= float32(0) || *dh.warnLevel > float32(1) {
		return &dh, fmt.Errorf("--warn must be > 0 and <= 1")
	}
	if *dh.critLevel <= float32(0) || *dh.critLevel > float32(1) {
		return &dh, fmt.Errorf("--crit must be > 0 and <= 1")
	}

	return &dh, nil
}

func (dh *DeploymentStatus) Usage() CheckUsage {
	d := "Checks deployment pod levels via status obj given by Kubernetes. Provides full deployment status object in result output\n"
	d += "Example: `deployment_status -w 0.8 -c 0.6`"
	return CheckUsage{
		Description: d,
		Flags: dh.commandLine.FlagUsages(),
	}
}

func (dh *DeploymentStatus) Update(resource interface{}) {
	dh.deployment = resource.(*v1beta1.Deployment)
}

func (dh *DeploymentStatus) Execute() (CheckResult, error) {
	res := NewCheckResultFromConfig(dh.Config)

	status := dh.deployment.Status

	res.Status = OK
	level := float32(status.AvailableReplicas) / float32(status.Replicas)
	if level < *dh.critLevel {
		res.Status = CRITICAL
	} else if level < *dh.warnLevel {
		res.Status = WARN
	}

	if *dh.minReplicas > 0 && *dh.minReplicas > *dh.deployment.Spec.Replicas {
		res.Status = CRITICAL
		res.Output = fmt.Sprintf("Replicas configured to %d\n", dh.deployment.Spec.Replicas)
	}

	o, _ := json.MarshalIndent(status, "", "  ")
	res.Output += string(o)

	return res, nil
}

// register factory
func init() {
	RegisterCheck("deployment_status", NewDeploymentStatus, []string{"deployment"})
}
