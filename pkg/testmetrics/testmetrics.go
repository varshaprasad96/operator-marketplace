package testmetrics

import (
	"context"
	"errors"
	"net/http"
	"time"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	MetricsEndPoint = ":8080"
)

var (
	opsProcessed = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events Test",
	})
	gaugeTest = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events Test",
	})

	metricsList = []prometheus.Collector{
		opsProcessed,
		gaugeTest,
	}
)

func RegisterMetrics() error {
	for _, metric := range metricsList {
		err := prometheus.Register(metric)
		if err != nil {
			return err
		}
	}
	return nil
}

// Custom metrics
func RecordMetrics() {
	//Register the metrics if required.
	//register()
	go func() {
		for {
			opsProcessed.Inc()
			gaugeTest.Inc()
			time.Sleep(2 * time.Second)
			//log.Printf("Recording custom metrics")
		}
	}()
}

// Register metrics, start recording and specify the endpoint
func StartMetrics() {
	RegisterMetrics()
	RecordMetrics()

	http.Handle("/metrics", prometheus.Handler())
	go http.ListenAndServe(MetricsEndPoint, nil)
}

// ErrMetricsFailedGenerateService indicates the metric service failed to generate
var ErrMetricsFailedGenerateService = errors.New("FailedGeneratingService")

// ErrMetricsFailedCreateService indicates that the service failed to create
var ErrMetricsFailedCreateService = errors.New("FailedCreateService")

// ErrMetricsFailedCreateRoute indicates that an account creation failed
var ErrMetricsFailedCreateRoute = errors.New("FailedCreateRoute")

// GenerateService returns the static service which exposes specifed port.
func GenerateService(port int32, portName string) (*v1.Service, error) {
	operatorName, err := k8sutil.GetOperatorName()
	if err != nil {
		return nil, err
	}
	namespace, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		return nil, err
	}

	label := map[string]string{"name": operatorName}

	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      operatorName,
			Namespace: namespace,
			Labels:    label,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port:     port,
					Protocol: v1.ProtocolTCP,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: port,
					},
					Name: portName,
				},
			},
			Selector: label,
		},
	}

	return service, nil
}

func createClient() (client.Client, error) {
	config, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	client, err := client.New(config, client.Options{})
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Creat or Update a service object which selects the pods from the operator which was deployed.
func createOrUpdateService(ctx context.Context, client client.Client, s *v1.Service) (*v1.Service, error) {
	if err := client.Create(ctx, s); err != nil {
		if !k8serr.IsAlreadyExists(err) {
			return nil, err
		}
		// Service already exists, we want to update it
		// as we do not know if any fields might have changed.
		existingService := &v1.Service{}
		err := client.Get(ctx, types.NamespacedName{
			Name:      s.Name,
			Namespace: s.Namespace,
		}, existingService)

		s.ResourceVersion = existingService.ResourceVersion
		if existingService.Spec.Type == v1.ServiceTypeClusterIP {
			s.Spec.ClusterIP = existingService.Spec.ClusterIP
		}
		err = client.Update(ctx, s)
		if err != nil {
			return nil, err
		}
		log.Printf("Metrics Service object updated Service.Name %v and Service.Namespace %v", s.Name, s.Namespace)
		return existingService, nil
	}

	log.Printf("Metrics Service object created Service.Name %v and Service.Namespace %v", s.Name, s.Namespace)
	return s, nil
}

// ConfigureMetrics generates metrics service and route,
// creates the metrics service and route,
// and finally it starts the metrics server
func ConfigureMetrics(ctx context.Context) error {
	log.Info("Starting prometheus metrics")

	// Start registering and recording metrics
	StartMetrics()

	client, err := createClient()
	if err != nil {
		log.Info("Failed to create new client", "Error", err.Error())
		return nil
	}

	// Generate Service Object
	s, svcerr := GenerateService(8080, "metrics")
	if svcerr != nil {
		log.Info("Error generating metrics service object.", "Error", svcerr.Error())
		return ErrMetricsFailedGenerateService
	}
	log.Info("Generated metrics service object")

	// Create or update Service
	_, err = createOrUpdateService(ctx, client, s)
	if err != nil {
		log.Info("Error getting current metrics service", "Error", err.Error())
		return ErrMetricsFailedCreateService
	}

	log.Info("Created Service")

	path := "/metrics"

	// Generate Route Object
	r := GenerateRoute(s, path)
	log.Info("Generated metrics route object")

	// Create or Update the Route
	err = client.Create(ctx, r)
	if err != nil {
		if k8serr.IsAlreadyExists(err) {
			// update the Route
			if rUpdateErr := client.Update(ctx, r); rUpdateErr != nil {
				log.Info("Error creating metrics route", "Error", rUpdateErr.Error())
				return ErrMetricsFailedCreateRoute
			}
			log.Info("Metrics route object updated", "Route.Name", r.Name, "Route.Namespace", r.Namespace)
			return nil
		}
		log.Info("Error creating metrics route", "Error", err.Error())
		return ErrMetricsFailedCreateRoute

	}
	log.Info("Metrics Route object Created", "Route.Name", r.Name, "Route.Namespace", r.Namespace)
	return nil
}

// Create route to the specified service.
func GenerateRoute(s *v1.Service, path string) *routev1.Route {
	log.Info("Staring to generate route modified")
	labels := make(map[string]string)
	for k, v := range s.ObjectMeta.Labels {
		labels[k] = v
	}

	return &routev1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: "route.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.ObjectMeta.Name,
			Namespace: s.ObjectMeta.Namespace,
			Labels:    labels,
		},
		Spec: routev1.RouteSpec{
			Path: path,
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: s.ObjectMeta.Name,
			},
			Port: &routev1.RoutePort{
				TargetPort: s.Spec.Ports[0].TargetPort,
			},
		},
	}
}

// Create service monitor based on the service which has been passed.
// Currently, Sservice monitor is not created, as routes are being used to expose
func GenerateServiceMonitor(s *v1.Service) *monitoringv1.ServiceMonitor {
	labels := make(map[string]string)
	for k, v := range s.ObjectMeta.Labels {
		labels[k] = v
	}

	return &monitoringv1.ServiceMonitor{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceMonitor",
			APIVersion: "monitoring.coreos.com/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.ObjectMeta.Name,
			Namespace: s.ObjectMeta.Namespace,
			Labels:    labels,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: labels,
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Port: s.Spec.Ports[0].Name,
				},
			},
		},
	}
}
