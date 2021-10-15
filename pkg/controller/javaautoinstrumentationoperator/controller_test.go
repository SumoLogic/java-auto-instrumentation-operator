package javaautoinstrumentationoperator

import (
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	"testing"
)

var testLogger = log.WithValues("Environment", "in test")

func TestShouldFindThatDeploymentNeedsInstrumentation(t *testing.T) {
	// given
	deployment := buildDeployment(map[string]string{
		"sumo-enable-instrumentation": "true",
	})

	// when
	needsInstrumentation := needsInstrumentation(deployment)

	// then
	assert.True(t, needsInstrumentation)
}

func TestShouldFindThatDeploymentHasAutoInstrumentationDisabled(t *testing.T) {
	// given
	deployment := buildDeployment(map[string]string{
		"sumo-enable-instrumentation": "false",
	})

	// when
	needsInstrumentation := needsInstrumentation(deployment)

	// then
	assert.False(t, needsInstrumentation)
}

func TestShouldFindThatDeploymentDoesntNeedAutoInstrumentation(t *testing.T) {
	// given
	deployment := buildDeployment(map[string]string{})

	// when
	needsInstrumentation := needsInstrumentation(deployment)

	// then
	assert.False(t, needsInstrumentation)
}

func TestShouldFindJavaOptionsWithOpenTelemetryForOneContainer(t *testing.T) {
	// given
	container := buildContainer("_JAVA_OPTIONS", "-javaagent:/jar/opentelemetry-javaagent-all.jar")

	// when
	hasAutoInstrJavaOpt := hasJavaOptionsEnvVarWithAutoInstrumentation([]corev1.Container{*container})

	// then
	assert.True(t, hasAutoInstrJavaOpt)
}

func TestShouldNotFindJavaOptionsWithOpenTelemetryForOneContainer(t *testing.T) {
	// given
	container := buildContainer("_JAVA_OPTIONS", "an option")

	// when
	hasAutoInstrJavaOpt := hasJavaOptionsEnvVarWithAutoInstrumentation([]corev1.Container{*container})

	// then
	assert.False(t, hasAutoInstrJavaOpt)
}

func TestShouldNotFindJavaOptionsForOneContainer(t *testing.T) {
	// given
	container := buildContainer("some env variable", "an option")

	// when
	hasAutoInstrJavaOpt := hasJavaOptionsEnvVarWithAutoInstrumentation([]corev1.Container{*container})

	// then
	assert.False(t, hasAutoInstrJavaOpt)
}

func TestShouldUseServiceNameFromLabel(t *testing.T) {
	// given
	deployment := buildDeployment(map[string]string{
		"sumo-service-name": "my cool service",
	})

	// when
	serviceName := getServiceName(testLogger, deployment)

	// then
	assert.Equal(t, "my cool service", serviceName)
}

func TestShouldUsePodHostnameAndContainerNameAsServiceName(t *testing.T) {
	// given
	deployment := buildDeployment(map[string]string{})

	// when
	serviceName := getServiceName(testLogger, deployment)

	// then
	// take a look at how the deployment is built
	assert.Equal(t, "podHost-container1", serviceName)
}

func TestShouldBuildJavaagentPath(t *testing.T) {
	// when
	path := getJavaagentPath()

	// then
	assert.Equal(t, " -javaagent:/jar/opentelemetry-javaagent-all.jar ", path)
}

func TestShouldCopyWithoutJavaOptions(t *testing.T) {
	// given
	envVars := []corev1.EnvVar{
		{Name: "E1", Value: "v1"},
		{Name: "_JAVA_OPTIONS", Value: "gc"},
		{Name: "E3", Value: "v3"},
	}

	// when
	copiedEnvVars := copyExistingEnvVarsWithoutJavaOptions(envVars)

	// then
	assert.Equal(t, 2, len(copiedEnvVars))
	assert.Equal(t, "E1", copiedEnvVars[0].Name)
	assert.Equal(t, "E3", copiedEnvVars[1].Name)
}

func TestShouldBuildOtJarsVolumeMount(t *testing.T) {
	// given
	volumeMount := getOtJarsVolumeMount()

	// expect
	assert.False(t, volumeMount.ReadOnly)
	assert.Equal(t, "sumo-ot-jar-volume", volumeMount.Name)
	assert.Equal(t, "/jar", volumeMount.MountPath)
}

func TestShouldBuildOtJarsVolume(t *testing.T) {
	// given
	volume := getOtJarsVolume()

	// expect
	assert.Equal(t, "sumo-ot-jar-volume", volume.Name)
}

func TestShouldEnableAutoInstrumentation(t *testing.T) {
	// given
	originalPodSpec := buildPodSpecForIntegration()
	serviceName := "my-cool-service"
	exporter := "jaeger"

	// when
	newPod := mergePodSpec(&originalPodSpec, serviceName, exporter, exporter)

	// then
	assert.Equal(t, 2, len(newPod.Volumes))
	assert.Equal(t, 2, len(newPod.Containers[0].Env))
}

func TestShouldFindCollectorHostFromDeploymentLabel(t *testing.T) {
	// given
	deployment := buildDeployment(map[string]string{
		"sumo-traces-collector-host": "ec2 machine",
	})

	// when
	host := getTracesCollectorHostOrDefault(deployment, "")

	// then
	assert.Equal(t, "ec2 machine", host)
}

func TestShouldUseExporterNameAsHostWhenNoLabelProvided(t *testing.T) {
	// given
	deployment := buildDeployment(map[string]string{})
	exporter := "my exporter"

	// when
	host := getTracesCollectorHostOrDefault(deployment, exporter)

	// then
	assert.Equal(t, exporter, host)
}

func TestShouldAddInitContainer(t *testing.T) {
	// given
	originalPodSpec := buildPodSpecForIntegration()
	assert.Equal(t, 0, len(originalPodSpec.InitContainers))

	// when
	newPodSpec := mergePodSpec(&originalPodSpec, "service", "exporter", "collector")

	// then
	assert.Equal(t, 1, len(newPodSpec.InitContainers))
}
