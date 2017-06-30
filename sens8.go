package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"flag"
	"strings"
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/informers"

	"github.com/hootsuite/sens8/client"
	"github.com/hootsuite/sens8/controller"
	"github.com/hootsuite/sens8/check"
	"github.com/hootsuite/sens8/util"
)

var (
	sensuConfigFile *string = flag.String("config-file", "/etc/sensu/config.json", "Sensu configuration file. Same format as Sensu proper")
	checkCommands *bool = flag.Bool("check-commands", false, "Print documentation for all check commands and exit")
	checkCommandsMd *bool = flag.Bool("check-commands-md", false, "Print documentation for all checks commands in markdown format and exit (indended for publishing docs)")
)

func main() {
	flag.Parse()

	if *checkCommands {
		printCheckCommandHelpText()
		os.Exit(0)
	}

	if *checkCommandsMd {
		printCheckCommandMarkdown()
		os.Exit(0)
	}

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
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(fmt.Sprintf("creating kubernetes client: %s", err.Error()))
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(fmt.Sprintf("creating kubernetes client: %s", err.Error()))
	}

	// init informers
	sharedInformers := informers.NewSharedInformerFactory(clientset, 300 * time.Second)

	// init controllers
	controllers := map[string]*controller.ResourceCheckController{}
	for _, res := range []string{"deployment", "pod", "daemonset"} {
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

func printCheckCommandHelpText() {
	for name, usage := range check.Docs() {
		fmt.Println(name)
		fmt.Println(util.PadRight("", "=", len(name)))
		fmt.Printf("Resources: %s\n", strings.Join(usage.Resources, ", "))
		fmt.Printf("%s\n\n", usage.Description)
		fmt.Printf("%s\n\n", usage.Flags)
	}
}

func printCheckCommandMarkdown() {
	fmt.Printf("Checks Commands\n===============\n\n")
	fmt.Printf("Get latest docs via: `./sens8 -check-commands`\n\n")

	for name, usage := range check.Docs() {
		fmt.Printf("### `%s`\n\n", name)
		fmt.Printf("**Resources**: %s\n\n", strings.Join(usage.Resources, ", "))
		fmt.Printf("%s\n\n", usage.Description)
		fmt.Printf("```\n%s\n```\n\n", usage.Flags)
	}
}

