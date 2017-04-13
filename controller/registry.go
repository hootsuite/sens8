package controller

import (
	"time"

	"github.com/golang/glog"

	"github.com/hootsuite/sens8/check"
	"github.com/hootsuite/sens8/client"
)

type CheckRegistry struct {
	resourceChecks map[string]ResourceChecks
	sensuClient    *client.SensuClient
}

type ResourceChecks map[uint64]CheckItem

type CheckItem struct {
	UpdateCh chan interface{}
}

// NewCheckRegistry creates a new CheckRegistry
func NewCheckRegistry(sensuClient *client.SensuClient) CheckRegistry {
	return CheckRegistry{
		resourceChecks: make(map[string]ResourceChecks),
		sensuClient: sensuClient,
	}
}

// Add adds checks to the registry and starts them
func (c *CheckRegistry) Add(checks []check.Check, resource interface{}, checkSource string) {

	if _, exists := c.resourceChecks[checkSource]; !exists {
		c.resourceChecks[checkSource] = make(ResourceChecks)
	}

	for _, check := range checks {
		// pre existing check, perhaps from hash collision. close and continue
		c.stopCheck(check, checkSource)

		updateCh := make(chan interface{})
		c.resourceChecks[checkSource][check.GetHash()] = CheckItem{
			UpdateCh: updateCh,
		}
		go c.startCheck(check, updateCh, resource)

		glog.V(1).Infof("%s %s: check added", checkSource, check.GetConfig().Name)
	}
}

// Update updates checks in the registry
func (c *CheckRegistry) Update(oldChecks []check.Check, newChecks []check.Check, resource interface{}, checkSource string) {
	// check old vs. new
	for _, o := range oldChecks {
		del := true
		// update existing check
		for _, n := range newChecks {
			if o.GetHash() == n.GetHash() {
				del = false
				checkItem, ok := c.resourceChecks[checkSource][n.GetHash()]
				if ok {
					glog.V(2).Infof("%s %s: updating resource", checkSource, o.GetConfig().Name)
					checkItem.UpdateCh <- resource
				}
				continue
			}
		}
		// delete old check
		if del {
			c.stopCheck(o, checkSource)
		}
	}

	// add new checks
	for _, n := range newChecks {
		if _, exists := c.resourceChecks[checkSource][n.GetHash()]; !exists {

			updateCh := make(chan interface{})
			c.resourceChecks[checkSource][n.GetHash()] = CheckItem{
				UpdateCh: updateCh,
			}
			go c.startCheck(n, updateCh, resource)

			glog.V(1).Infof("%s %s: check added", checkSource, n.GetConfig().Name)
		}
	}
}

// Delete stop checks and removes if from the the registry
func (c *CheckRegistry) Delete(checks []check.Check, resource interface{}, checkSource string) {
	if _, exists := c.resourceChecks[checkSource]; !exists {
		return
	}

	for _, check := range checks {
		c.stopCheck(check, checkSource)
	}
}

// startCheck starts the check. meant to be run in a goroutine
func (c *CheckRegistry) startCheck(check check.Check, updateCh chan interface{}, resource interface{}) {

	exec := func() {
		glog.V(5).Infof("%s %s: running check", *check.GetConfig().Source, check.GetConfig().Name)

		res, err := check.Execute()
		if err != nil {
			glog.Errorf("%s %s: error running check: %s", *check.GetConfig().Source, check.GetConfig().Name, err.Error())
			return
		}

		if err = c.sensuClient.PostCheckResult(res); err != nil {
			glog.Errorf("%s %s: error sending check result: %s", *check.GetConfig().Source, check.GetConfig().Name, err.Error())
		}
	}

	// kick off first check run
	ticker := time.NewTicker(time.Duration(check.GetConfig().Interval) * time.Second)
	check.Update(resource)
	exec()

	for {
		select {
		case d, ok := <-updateCh:
			if !ok {
				updateCh = nil
				ticker.Stop()
				glog.V(1).Infof("%s %s: check stopped", *check.GetConfig().Source, check.GetConfig().Name)
				return
			}
			check.Update(d)
		case <-ticker.C:
			exec()
		}
	}
}

// stopCheck stops the check goroutine
func (c *CheckRegistry) stopCheck(check check.Check, checkSource string) {
	if existing, exists := c.resourceChecks[checkSource][check.GetHash()]; exists {
		glog.V(1).Infof("%s %s: stopping check", checkSource, check.GetConfig().Name)
		close(existing.UpdateCh)
		delete(c.resourceChecks[checkSource], check.GetHash())
	}
}
