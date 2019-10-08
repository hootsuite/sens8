package service

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/glog"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/hootsuite/sens8/internal/checks"
	"github.com/hootsuite/sens8/internal/controller"
	"github.com/hootsuite/sens8/sensu"
)

var (
	sensuConfigFile *string = flag.String("config-file", "/etc/sensu/config.json", "Sensu configuration file. Same format as Sensu proper")
	checkDocs       *bool   = flag.Bool("check-docs", false, "Print documentation for all check commands and exit")
	checkDocsMd     *bool   = flag.Bool("check-docs-md", false, "Print documentation for all checks commands in markdown format and exit (indended for publishing docs)")
)

func main() {
	flag.Parse()

	if *checkDocs {
		fmt.Println(checks.GenCheckDocsText())
		os.Exit(0)
	}

	if *checkDocsMd {
		fmt.Println(checks.GenCheckDocsMarkdown())
		os.Exit(0)
	}

	// global stop channel - for all controllers & informers
	stopCh := make(chan struct{})

	// init sensu
	sensuClient, err := sensu.New(*sensuConfigFile)
	if err != nil {
		panic(fmt.Sprintf("creating sensu client: %s", err.Error()))
	}
	go sensuClient.Start(stopCh)
	checks.Defaults = sensuClient.Config.Defaults

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
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(fmt.Sprintf("creating kubernetes client: %s", err.Error()))
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(fmt.Sprintf("creating kubernetes client: %s", err.Error()))
	}

	// init informers
	sharedInformers := informers.NewSharedInformerFactory(clientset, 300*time.Second)

	// init controllers
	controllers := map[string]*controller.ResourceCheckController{}
	for _, res := range []string{"deployment", "pod", "daemonset", "service"} {
		ctl := controller.NewResourceCheckController(
			clientset,
			&sensuClient,
			controller.ResourceAdapterFactory(res, sharedInformers),
		)
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
