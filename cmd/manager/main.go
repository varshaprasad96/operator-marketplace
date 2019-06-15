package main

import (
	"flag"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	olm "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/signals"
	"github.com/operator-framework/operator-marketplace/pkg/apis"
	"github.com/operator-framework/operator-marketplace/pkg/catalogsourceconfig"
	"github.com/operator-framework/operator-marketplace/pkg/controller"
	"github.com/operator-framework/operator-marketplace/pkg/operatorsource"
	"github.com/operator-framework/operator-marketplace/pkg/registry"
	"github.com/operator-framework/operator-marketplace/pkg/status"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"github.com/operator-framework/operator-sdk/pkg/restmapper"
	sdkVersion "github.com/operator-framework/operator-sdk/version"

	log "github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func recordMetrics() {
	go func() {
		for {
			opsProcessed.Inc()
			time.Sleep(2 * time.Second)
			log.Printf("Recording custom metrics")
		}
	}()
}

var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events Test",
	})
)

const (
	// TODO: resyncInterval is hardcoded to 1 hour now, it would have to be
	// configurable on a per OperatorSource level.
	resyncInterval = time.Duration(60) * time.Minute

	initialWait                = time.Duration(1) * time.Minute
	updateNotificationSendWait = time.Duration(10) * time.Minute
)

func printVersion() {
	log.Printf("Hello Go Version: %s", runtime.Version())
	log.Printf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	log.Printf("operator-sdk Version: %v", sdkVersion.Version)
}

// func init() {
// 	prometheus.MustRegister(opsProcessed)
// 	log.Printf("Registered")
// }

func main() {
	printVersion()
	recordMetrics()

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":8080", nil)

	log.Info(" Custom Metircs recorded")

	// Parse the command line arguments for the registry server image
	flag.StringVar(&registry.ServerImage, "registryServerImage",
		registry.DefaultServerImage, "the image to use for creating the operator registry pod")
	flag.Parse()

	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		log.Fatalf("failed to get watch namespace: %v", err)
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	// // Create a new Cmd to provide shared dependencies and start components
	// mgr, err := manager.New(cfg, manager.Options{Namespace: namespace})
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// Change the below variables to serve metrics on different host or port.
	// var metricsHost = "0.0.0.0"
	// var metricsPort int32 = 8383
	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		Namespace:      namespace,
		MapperProvider: restmapper.NewDynamicRESTMapper,
		//MetricsBindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort),
	})
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	log.Print("Registering Components.")

	catalogsourceconfig.InitializeStaticSyncer(mgr.GetClient(), initialWait)
	registrySyncer := operatorsource.NewRegistrySyncer(mgr.GetClient(), initialWait, resyncInterval, updateNotificationSendWait, catalogsourceconfig.Syncer, catalogsourceconfig.Syncer)

	// Setup Scheme for all defined resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		exit(err)
	}

	// Add external resource to scheme
	if err := olm.AddToScheme(mgr.GetScheme()); err != nil {
		exit(err)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		exit(err)
	}

	//ctx := context.TODO()

	// Create Service object to expose the metrics port.
	/*var metricsPort int32 = 2112
	service, err := metrics.ExposeMetricsPort(ctx, metricsPort)
	fmt.Printf("Service %v", service)
	if err != nil {
		log.Info("ERROR EXPOSING METRICS")
		log.Info(err.Error())
	}*/
	// Create one `ServiceMonitor` per application per namespace.
	// Change below value to name of the Namespace you want the `ServiceMonitor` to be created in.
	/*ns := "openshift-marketplace"
	// Populate below with the Service(s) for which you want to create ServiceMonitors.
	services := []*v1.Service{}
	services = append(services, service)

	fmt.Printf("Services %v", services)

	// Pass the Service(s) to the helper function, which in turn returns the array of `ServiceMonitor` objects.
	_, err = metrics.CreateServiceMonitors(cfg, ns, services)
	if err != nil {
		log.Info("ERROR CREATING SERVICE MONITORS")
		log.Info(err.Error())
	} else {
		log.Info("NO ERROR CREATING SERVICE MONITORS")

	}

	log.Info("Expose-Metrics succeded")
	log.Info("Starting the Cmd.")*/

	// Serve a health check.
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	go http.ListenAndServe(":8080", nil)

	log.Print("Starting the Cmd.")
	stopCh := signals.SetupSignalHandler()

	// statusReportingDoneCh will be closed after the operator has successfully stopped reporting ClusterOperator status.
	statusReportingDoneCh := status.StartReporting(cfg, mgr, namespace, os.Getenv("RELEASE_VERSION"), stopCh)

	go registrySyncer.Sync(stopCh)
	go catalogsourceconfig.Syncer.Sync(stopCh)

	// Start the Cmd
	err = mgr.Start(stopCh)

	// Wait for ClusterOperator status reporting routine to close the statusReportingDoneCh channel.
	<-statusReportingDoneCh

	exit(err)
}

// exit stops the reporting of ClusterOperator status and exits with the proper exit code.
func exit(err error) {
	// If an error exists then exit with status set to 1
	if err != nil {
		log.Fatalf("The operator encountered an error, exit code 1: %v", err)
	}

	// No error, graceful termination
	log.Info("The operator exited gracefully, exit code 0")
	os.Exit(0)
}
