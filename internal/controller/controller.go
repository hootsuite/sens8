package controller

import (
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/hootsuite/sens8/sensu"
	"github.com/hootsuite/sens8/internal/checks"
)

const (
	CheckAnnotation = "hootsuite.com/sensu-checks"
)

type ResourceCheckController struct {
	// resource adapter
	adapter ResourceAdapter
	// kubernetes client
	clientset kubernetes.Interface
	// store for all the instantiated checks
	registry CheckRegistry
	// sensu client
	sensuClient *sensu.Client
}

// NewResourceCheckController creates a new controller for k8s resources based on what adapter is passed in.
func NewResourceCheckController(clientset kubernetes.Interface, sensuClient *sensu.Client, adapter ResourceAdapter) ResourceCheckController {

	c := ResourceCheckController{
		adapter:     adapter,
		clientset:   clientset,
		registry:    NewCheckRegistry(sensuClient),
		sensuClient: sensuClient,
	}

	adapter.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addResource,
		UpdateFunc: c.updateResource,
		DeleteFunc: c.deleteResource,
	})
	return c
}

func (c *ResourceCheckController) Run(stopCh chan struct{}) {
	glog.Infof("starting %s controller", c.adapter.Type())

	if !cache.WaitForCacheSync(stopCh, c.adapter.Informer().HasSynced) {
		return
	}

	<-stopCh
	glog.Infof("shutting down %s controller", c.adapter.Type())
}

// addResource is an informer handler
func (c *ResourceCheckController) addResource(obj interface{}) {
	glog.V(6).Infof("%s.addResource", c.adapter.Type())
	checkSource := c.adapter.CheckSource(obj)

	checks := c.getChecks(obj, checkSource)
	if len(checks) == 0 {
		return
	}
	c.registry.Add(checks, obj, checkSource)
}

// updateResource is an informer handler
func (c *ResourceCheckController) updateResource(oldObj, newObj interface{}) {
	glog.V(6).Infof("%s.updateResource", c.adapter.Type())
	checkSource := c.adapter.CheckSource(newObj)

	oldChecks := c.getChecks(oldObj, checkSource)
	newChecks := c.getChecks(newObj, checkSource)

	if len(newChecks) == 0 && len(oldChecks) == 0 {
		return
	}
	c.registry.Update(oldChecks, newChecks, newObj, checkSource)

	if len(newChecks) == 0 {
		c.deregister(oldChecks, checkSource)
	}
}

// deleteResource is an informer handler
func (c *ResourceCheckController) deleteResource(obj interface{}) {
	glog.V(6).Infof("%s.deleteResource", c.adapter.Type())
	checkSource := c.adapter.CheckSource(obj)

	checks := c.getChecks(obj, checkSource)
	if len(checks) == 0 {
		return
	}
	c.registry.Delete(checks, obj, checkSource)

	c.deregister(checks, checkSource)
}

// deregister determines and acts if the client (source) should be de-registered in sensu
func (c *ResourceCheckController) deregister(checks []checks.Check, checkSource string) {
	keep := false
	useDefault := true
	for _, check := range checks {
		d := check.GetConfig().Deregister
		if d != nil {
			keep = keep || !*d
			useDefault = false
		}
	}
	if useDefault {
		keep = !c.adapter.DeregisterDefault()
	}
	if !keep {
		glog.V(1).Infof("deregistering client %s in sensu", checkSource)
		if err := c.sensuClient.Deregister(checkSource); err != nil {
			glog.Errorf("error deregistering client %s in sensu: %s", checkSource, err.Error())
		}
	}
}

// getConfigs extracts and filters checks from the resource's annotation meta
func (c *ResourceCheckController) getChecks(resource interface{}, checkSource string) []checks.Check {
	checkConfigs, ok := c.adapter.CheckConfigs(resource)
	if !ok {
		return []checks.Check{}
	}

	checks, errors := checks.ParseCheckConfigs(checkConfigs, checkSource, c.adapter.Type())
	for _, err := range errors {
		glog.Errorf("%s: %s", checkSource, err.Error())
	}
	return checks
}
