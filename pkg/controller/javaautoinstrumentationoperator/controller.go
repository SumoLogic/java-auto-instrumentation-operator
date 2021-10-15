package javaautoinstrumentationoperator

import (
	"context"
	"strings"
	"time"

	"github.com/go-logr/logr"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	javaautoinstrv1alpha1 "github.com/SumoLogic/java-auto-instrumentation-operator/pkg/apis/javaautoinstr/v1alpha1"
)

// Labels
const enableInstrumentationLabel = "sumo-enable-instrumentation"
const serviceNameLabel = "sumo-service-name"
const opentelemetryTracesExporterLabel = "sumo-traces-exporter"
const opentelemetryTracesCollectorHostLabel = "sumo-traces-collector-host"

// Docker images stuff
const opentelemetryJarContainerName = "sumo-ot-jar-holder"
const opentelemetryJarImgTag = "1.6.2"
const opentelemetryJarContainerImage = "public.ecr.aws/a4t4y2n3/opentelemetry-java-instrumentation-jar:" + opentelemetryJarImgTag

const opentelemetryJarVolumeName = "sumo-ot-jar-volume"

const opentelemetryJavaagentJarName = "opentelemetry-javaagent-all.jar"
const opentelemetryJarPath = "/ot-jar/" + opentelemetryJavaagentJarName
const opentelemetryJarMountPath = "/jar"
const opentelemetryJarFinalPath = opentelemetryJarMountPath + "/" + opentelemetryJavaagentJarName

const javaOptionsEnvVar = "_JAVA_OPTIONS"

var log = logf.Log.WithName("controller_javaautoinstrumentation")

// ReconcileJavaAutoInstrumentation reconciles a JavaAutoInstrumentation object
type ReconcileJavaAutoInstrumentation struct {
	client.Client
	Scheme *runtime.Scheme
}

// Add creates a new JavaAutoInstrumentation Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileJavaAutoInstrumentation{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("javaautoinstrumentationoperator-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource JavaAutoInstrumentation
	err = c.Watch(&source.Kind{Type: &javaautoinstrv1alpha1.JavaAutoInstrumentation{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner JavaAutoInstrumentation
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &javaautoinstrv1alpha1.JavaAutoInstrumentation{},
	})
	if err != nil {
		return err
	}

	log.Info("Watching all deployments")
	err = c.Watch(&source.Kind{Type: &appv1.Deployment{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		log.Error(err, "Failed to watch all deployments")
		return err
	}
	return nil
}

// blank assignment to verify that ReconcileJavaAutoInstrumentation implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileJavaAutoInstrumentation{}

// Reconcile reads that state of the cluster for a JavaAutoInstrumentation object and makes changes based on the state read
// and what is in the JavaAutoInstrumentation.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileJavaAutoInstrumentation) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	now := time.Now().Format(time.RFC3339)
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name,
		"Timestamp", now)
	reqLogger.Info("Reconciling JavaAutoInstrumentation")

	existingDeployments := &appv1.DeploymentList{}
	err := r.Client.List(context.TODO(), existingDeployments, &client.ListOptions{Namespace: request.Namespace})
	if err != nil {
		reqLogger.Error(err, "failed to list existing deployments")
		return reconcile.Result{}, err
	}

	for _, deployment := range existingDeployments.Items {
		reqLogger.Info("Processing", "deployment", deployment.Name)
		if needsInstrumentation(&deployment) {
			if !hasJavaOptionsEnvVarWithAutoInstrumentation(deployment.Spec.Template.Spec.Containers) {
				reqLogger.Info("Containers do not have _JAVA_OPTIONS env var with auto instrumentation",
					"Deployment", deployment.Name)
				tracesExporter := getTracesExporterOrDefault(reqLogger, &deployment)
				tracesCollectorHost := getTracesCollectorHostOrDefault(&deployment, tracesExporter)
				serviceName := getServiceName(reqLogger, &deployment)
				deployment.Spec.Template.Spec = mergePodSpec(&deployment.Spec.Template.Spec, serviceName,
					tracesExporter, tracesCollectorHost)
				err = r.Client.Update(context.TODO(), &deployment)
				if err != nil {
					reqLogger.Error(err, "Failed to update deployment", "Deployment", deployment.Name)
					return reconcile.Result{}, err
				} else {
					reqLogger.Info("Successfully updated deployment", "Deployment", deployment.Name)
				}
			} else {
				reqLogger.Info("Containers have _JAVA_OPTIONS with auto instrumentation, will leave them alone")
			}
		} else {
			reqLogger.Info("This deployment doesn't need auto instrumentation")
		}
	}
	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReconcileJavaAutoInstrumentation) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&javaautoinstrv1alpha1.JavaAutoInstrumentation{}).
		Owns(&appv1.Deployment{}).
		Complete(r)
}

func getTracesCollectorHostOrDefault(deployment *appv1.Deployment, exporter string) string {
	providedHost, ok := deployment.Labels[opentelemetryTracesCollectorHostLabel]
	if ok {
		return providedHost
	} else {
		return exporter
	}
}

func hasJavaOptionsEnvVarWithAutoInstrumentation(containers []corev1.Container) bool {
	options, exists := getJavaOptions(containers)
	return exists && strings.Contains(options, opentelemetryJavaagentJarName)
}

func getJavaOptions(containers []corev1.Container) (string, bool) {
	for _, container := range containers {
		for _, e := range container.Env {
			if e.Name == javaOptionsEnvVar {
				return e.Value, true
			}
		}
	}
	return "", false
}

func needsInstrumentation(deployment *appv1.Deployment) bool {
	enableInstrumentation, ok := deployment.Labels[enableInstrumentationLabel]
	if ok {
		return enableInstrumentation == "true"
	}
	return false
}

func getServiceName(reqLogger logr.Logger, deployment *appv1.Deployment) string {
	name, ok := deployment.Labels[serviceNameLabel]
	if ok {
		reqLogger.Info("Using label for tracing service name")
		return name
	} else {
		podSpec := deployment.Spec.Template.Spec
		numberOfContainers := len(podSpec.Containers)
		reqLogger.Info("Using pod container for tracing service name", "Number of containers",
			numberOfContainers)
		return podSpec.Hostname + "-" + podSpec.Containers[0].Name
	}
}

func getConfiguration(exporter string, serviceName string, existingJavaOptions string,
	collectorHost string) []corev1.EnvVar {

	if exporter == "zipkin" {
		return getZipkinConfiguration(existingJavaOptions, collectorHost, serviceName)

	} else if exporter == "jaeger" {
		return getJaegerConfiguration(existingJavaOptions, collectorHost, serviceName)
	} else {
		return getTracesOtlpConfiguration(existingJavaOptions, collectorHost, serviceName)
	}
}

func copyExistingEnvVarsWithoutJavaOptions(env []corev1.EnvVar) []corev1.EnvVar {
	var envVars []corev1.EnvVar
	for _, e := range env {
		if e.Name != javaOptionsEnvVar {
			envVars = append(envVars, e)
		}
	}
	return envVars
}
