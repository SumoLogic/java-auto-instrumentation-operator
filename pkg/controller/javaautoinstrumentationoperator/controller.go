package javaautoinstrumentationoperator

import (
	"context"
	"github.com/go-logr/logr"
	"strings"
	"time"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	javaautoinstrv1alpha1 "github.com/SumoLogic/java-auto-instrumentation-operator/pkg/apis/javaautoinstr/v1alpha1"
)

const enableInstrumentationLabel = "sumo-enable-instrumentation"
const serviceNameLabel = "sumo-service-name"

const opentelemetryTracesExporterLabel = "sumo-traces-exporter"
const opentelemetryTracesCollectorHostLabel = "sumo-traces-collector-host"

const opentelemetryJarVolumeName = "sumo-ot-jar-volume"
const opentelemetryJarMountPath = "/ot-jar"

const opentelemetryJarContainerName = "sumo-ot-jar-holder"

const opentelemetryJarVersion = "1.6.2"
const opentelemetryJarContainerImage = "public.ecr.aws/a4t4y2n3/opentelemetry-java-instrumentation-jar:" + opentelemetryJarVersion
const opentelemetryJavaagentJarName = "opentelemetry-javaagent-all.jar"

var log = logf.Log.WithName("controller_javaautoinstrumentation")

// Add creates a new JavaAutoInstrumentation Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileJavaAutoInstrumentation{client: mgr.GetClient(), scheme: mgr.GetScheme()}
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

// ReconcileJavaAutoInstrumentation reconciles a JavaAutoInstrumentation object
type ReconcileJavaAutoInstrumentation struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a JavaAutoInstrumentation object and makes changes based on the state read
// and what is in the JavaAutoInstrumentation.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileJavaAutoInstrumentation) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	now := time.Now().Format(time.RFC3339)
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name,
		"Timestamp", now)
	reqLogger.Info("Reconciling JavaAutoInstrumentation")

	existingDeployments := &appv1.DeploymentList{}
	err := r.client.List(context.TODO(), existingDeployments, &client.ListOptions{Namespace: request.Namespace})
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
				err = r.client.Update(context.TODO(), &deployment)
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
	return exists && strings.Contains(options, "opentelemetry-auto")
}

func getJavaOptions(containers []corev1.Container) (string, bool) {
	for _, container := range containers {
		for _, e := range container.Env {
			if e.Name == "_JAVA_OPTIONS" {
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

func getTracesExporterOrDefault(reqLogger logr.Logger, deployment *appv1.Deployment) string {
	exporter, ok := deployment.Labels[opentelemetryTracesExporterLabel]
	if ok {
		if exporter == "jaeger" || exporter == "otlp" || exporter == "zipkin" {
			return exporter
		} else {
			reqLogger.Info("Unknown exporter "+exporter+", will default to OTLP/HTTP protobuf", "Deployment",
				deployment.Name)
			return "otlp"
		}
	} else {
		reqLogger.Info("No exporter set, will default to OTLP/HTTP protobuf", "Deployment",
			deployment.Name)
		return "otlp"
	}
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

func getJavaagentPath() string {
	return " -javaagent:\"" + opentelemetryJarMountPath + "/" + opentelemetryJavaagentJarName + "\" "
}

func getJaegerConfiguration(existingJavaOptions string, collectorHost string, serviceName string) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name: "_JAVA_OPTIONS",
			Value: existingJavaOptions + getJavaagentPath() +
				"-Dotel.traces.exporter=jaeger " +
				"-Dotel.exporter.jaeger.endpoint=http://" + collectorHost + ":14250 " +
				"-Dotel.service.name=" + serviceName + " ",
		},
	}
}

func getZipkinConfiguration(existingJavaOptions string, collectorHost string, serviceName string) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name: "_JAVA_OPTIONS",
			Value: existingJavaOptions + getJavaagentPath() +
				"-Dotel.traces.exporter=zipkin " +
				"-Dotel.exporter.zipkin.endpoint=http://" + collectorHost + ":9411/api/v2/spans " +
				"-Dotel.service.name=" + serviceName + " ",
		},
	}
}

func getTracesOtlpConfiguration(existingJavaOptions string, collectorHost string, serviceName string) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name: "_JAVA_OPTIONS",
			Value: existingJavaOptions + getJavaagentPath() +
				"-Dotel.traces.exporter=otlp " +
				"-Dotel.exporter.otlp.traces.endpoint=http://" + collectorHost + ":55681/v1/traces " +
				"-Dotel.service.name=" + serviceName + " " +
				"-Dotel.exporter.otlp.protocol=http/protobuf ",
		},
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
		if e.Name != "_JAVA_OPTIONS" {
			envVars = append(envVars, e)
		}
	}
	return envVars
}

func getOtJarsVolumeMount() corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      opentelemetryJarVolumeName,
		MountPath: opentelemetryJarMountPath,
	}
}

func getOtJarsVolume() corev1.Volume {
	return corev1.Volume{
		Name: opentelemetryJarVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
}

func mergePodSpec(originalPodSpec *corev1.PodSpec, serviceName string, exporter string,
	collectorHost string) corev1.PodSpec {

	originalContainer := originalPodSpec.Containers[0] // TODO
	existingJavaOptions, exists := getJavaOptions(originalPodSpec.Containers)
	if !exists {
		existingJavaOptions = ""
	}
	var envVars = copyExistingEnvVarsWithoutJavaOptions(originalContainer.Env)
	envVars = append(envVars, getConfiguration(exporter, serviceName, existingJavaOptions, collectorHost)...)

	otJarsVolumeMount := getOtJarsVolumeMount()
	var volumeMounts []corev1.VolumeMount
	volumeMounts = append(volumeMounts, originalContainer.VolumeMounts...)
	volumeMounts = append(volumeMounts, otJarsVolumeMount)

	var volumes []corev1.Volume
	volumes = append(volumes, originalPodSpec.Volumes...)
	volumes = append(volumes, getOtJarsVolume())

	initContainers := append(originalPodSpec.InitContainers, getOtJarsInitContainer(otJarsVolumeMount))

	return corev1.PodSpec{
		Volumes:        volumes,
		InitContainers: initContainers,
		Containers: []corev1.Container{
			{
				Name:                     originalContainer.Name,
				Image:                    originalContainer.Image,
				Resources:                originalContainer.Resources,
				SecurityContext:          originalContainer.SecurityContext,
				Env:                      envVars,
				VolumeMounts:             volumeMounts,
				Command:                  originalContainer.Command,
				Args:                     originalContainer.Args,
				WorkingDir:               originalContainer.WorkingDir,
				Ports:                    originalContainer.Ports,
				EnvFrom:                  originalContainer.EnvFrom,
				VolumeDevices:            originalContainer.VolumeDevices,
				LivenessProbe:            originalContainer.LivenessProbe,
				ReadinessProbe:           originalContainer.ReadinessProbe,
				StartupProbe:             originalContainer.StartupProbe,
				Lifecycle:                originalContainer.Lifecycle,
				TerminationMessagePath:   originalContainer.TerminationMessagePath,
				TerminationMessagePolicy: originalContainer.TerminationMessagePolicy,
				ImagePullPolicy:          originalContainer.ImagePullPolicy,
				Stdin:                    originalContainer.Stdin,
				StdinOnce:                originalContainer.StdinOnce,
				TTY:                      originalContainer.TTY,
			},
		},
		EphemeralContainers:           originalPodSpec.EphemeralContainers,
		RestartPolicy:                 originalPodSpec.RestartPolicy,
		TerminationGracePeriodSeconds: originalPodSpec.TerminationGracePeriodSeconds,
		ActiveDeadlineSeconds:         originalPodSpec.ActiveDeadlineSeconds,
		DNSPolicy:                     originalPodSpec.DNSPolicy,
		NodeSelector:                  originalPodSpec.NodeSelector,
		ServiceAccountName:            originalPodSpec.ServiceAccountName,
		DeprecatedServiceAccount:      originalPodSpec.DeprecatedServiceAccount,
		AutomountServiceAccountToken:  originalPodSpec.AutomountServiceAccountToken,
		NodeName:                      originalPodSpec.NodeName,
		HostNetwork:                   originalPodSpec.HostNetwork,
		HostPID:                       originalPodSpec.HostPID,
		HostIPC:                       originalPodSpec.HostIPC,
		ShareProcessNamespace:         originalPodSpec.ShareProcessNamespace,
		SecurityContext:               originalPodSpec.SecurityContext,
		ImagePullSecrets:              originalPodSpec.ImagePullSecrets,
		Hostname:                      originalPodSpec.Hostname,
		Subdomain:                     originalPodSpec.Subdomain,
		Affinity:                      originalPodSpec.Affinity,
		SchedulerName:                 originalPodSpec.SchedulerName,
		Tolerations:                   originalPodSpec.Tolerations,
		HostAliases:                   originalPodSpec.HostAliases,
		PriorityClassName:             originalPodSpec.PriorityClassName,
		Priority:                      originalPodSpec.Priority,
		DNSConfig:                     originalPodSpec.DNSConfig,
		ReadinessGates:                originalPodSpec.ReadinessGates,
		RuntimeClassName:              originalPodSpec.RuntimeClassName,
		EnableServiceLinks:            originalPodSpec.EnableServiceLinks,
		PreemptionPolicy:              originalPodSpec.PreemptionPolicy,
		Overhead:                      originalPodSpec.Overhead,
		TopologySpreadConstraints:     originalPodSpec.TopologySpreadConstraints,
	}
}

func getOtJarsInitContainer(volumeMount corev1.VolumeMount) corev1.Container {
	return corev1.Container{
		Name:         opentelemetryJarContainerName,
		Image:        opentelemetryJarContainerImage,
		VolumeMounts: []corev1.VolumeMount{volumeMount},
	}
}
