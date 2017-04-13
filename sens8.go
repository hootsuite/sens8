package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"flag"
	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/controller/informers"

	"github.com/hootsuite/sens8/client"
	"github.com/hootsuite/sens8/controller"
	"github.com/hootsuite/sens8/check"
)

var (
	sensuConfigFile *string = flag.String("config-file", "/etc/sensu/config.json", "Sensu configuration file. Same format as Sensu proper")
)

func main() {
	flag.Parse()

	// global stop channel - for all controllers & informers
	stopCh := make(chan struct{})

	// init sensu
	sensuClient, err := client.NewSensuClient(*sensuConfigFile)
	if err != nil {
		panic(fmt.Sprintf("creating sensu client: %s", err.Error()))
	}
	go sensuClient.Start(stopCh)
	check.Defaults = sensuClient.Config.Defaults

	// give some time for client to connect for first batch of results to send out
	glog.Info("waiting for rabbitmq connection")
	for i := 0; i < 50; i++ {
		if sensuClient.Transport.IsConnected() {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	// kick off keepalive check for this service
	go sensuClient.StartKeepalive(stopCh)


	// Set up the kubernetes go clients
	config, err := restclient.InClusterConfig()
	if err != nil {
		panic(fmt.Sprintf("creating kubernetes client: %s", err.Error()))
	}
	clientset, err := internalclientset.NewForConfig(config)
	if err != nil {
		panic(fmt.Sprintf("creating kubernetes client: %s", err.Error()))
	}

	// init informers
	sharedInformers := informers.NewSharedInformerFactory(clientset, 300 * time.Second)

	// init controllers
	controllers := map[string]*controller.ResourceCheckController{}
	for _, res := range []string{"deployment", "pod"} {
		ctl, err := controller.NewResourceCheckController(
			clientset,
			&sensuClient,
			controller.ResourceAdapterFactory(res, sharedInformers),
		)
		if err != nil {
			panic(fmt.Sprintf("failed to create %s controller: %s", res, err.Error()))
		}
		controllers[res] = &ctl
		go ctl.Run(stopCh)
	}

	// kick off informers
	sharedInformers.Start(stopCh)

	// block & close cleanly
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	<-sig

	glog.Info("Caught signal. Shutting down")
	close(stopCh)

	// give gorotines listening to stopCh a chance to shutdown.
	// we might be flushing out remaining results
	time.Sleep(2 * time.Second)
	glog.Flush()
}
