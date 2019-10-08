package checks

import (
	"encoding/json"
	"fmt"

	flag "github.com/spf13/pflag"
	apps_v1 "k8s.io/api/apps/v1"
)

type DaemonSetStatus struct {
	BaseCheck
	warnLevel   *float32
	critLevel   *float32
	daemonSet   *apps_v1.DaemonSet
	commandLine *flag.FlagSet
}

//NewDaemonSetStatus creates a new daemonSet health check
func NewDaemonSetStatus(config CheckConfig) (Check, error) {
	dh := DaemonSetStatus{}
	dh.Config = config

	// process flags
	commandLine := flag.NewFlagSet(config.Id, flag.ContinueOnError)
	dh.warnLevel = commandLine.Float32P("warn", "w", 1.0, "Percent of healthy (available) pods to warn at")
	dh.critLevel = commandLine.Float32P("crit", "c", 0.9, "Percent of healthy (available) pods to alert critical at")
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

func (dh *DaemonSetStatus) Usage() CheckUsage {
	d := "Checks daemonSet pod levels via status obj given by Kubernetes. Provides full daemonSet status object in result output\n"
	d += "Example: `daemonset_status -w 1.0 -c 0.9`"
	return CheckUsage{
		Description: d,
		Flags:       dh.commandLine.FlagUsages(),
	}
}

func (dh *DaemonSetStatus) Update(resource interface{}) {
	dh.daemonSet = resource.(*apps_v1.DaemonSet)
}

func (dh *DaemonSetStatus) Execute() (CheckResult, error) {
	res := NewCheckResultFromConfig(dh.Config)

	status := dh.daemonSet.Status

	res.Status = OK
	level := float32(status.NumberAvailable) / float32(status.DesiredNumberScheduled)
	if level < *dh.critLevel {
		res.Status = CRITICAL
	} else if level < *dh.warnLevel {
		res.Status = WARN
	}

	o, _ := json.MarshalIndent(status, "", "  ")
	res.Output += string(o)

	return res, nil
}

// register factory
func init() {
	RegisterCheck("daemonset_status", NewDaemonSetStatus, []string{"daemonset"})
}
