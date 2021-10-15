package javaautoinstrumentationoperator

import (
	"strings"

	"github.com/stretchr/testify/assert"

	"testing"
)

func TestShouldUseDefaultExporterIfNoExplicitOneFound(t *testing.T) {
	// given
	deployment := buildDeployment(map[string]string{})

	// when
	exporter := getTracesExporterOrDefault(testLogger, deployment)

	// then
	assert.Equal(t, "otlp", exporter)
}

func TestShouldUseOtlpExporterExplicitly(t *testing.T) {
	// given
	deployment := buildDeployment(map[string]string{
		"sumo-traces-exporter": "otlp",
	})

	// when
	exporter := getTracesExporterOrDefault(testLogger, deployment)

	// then
	assert.Equal(t, "otlp", exporter)
}

func TestShouldFallbackToOtlpExporterWhenExporterIsUnknown(t *testing.T) {
	// given
	deployment := buildDeployment(map[string]string{
		"sumo-traces-exporter": "unknown exporter",
	})

	// when
	exporter := getTracesExporterOrDefault(testLogger, deployment)

	// then
	assert.Equal(t, "otlp", exporter)
}

func TestShouldBuildOtlpConfiguration(t *testing.T) {
	// given
	serviceName := "super-app"
	existingOpts := "some-opts"
	collectorHost := "otlp-host"

	// when
	config := getTracesOtlpConfiguration(existingOpts, collectorHost, serviceName)

	// then
	assert.Equal(t, 1, len(config))

	assert.Equal(t, "_JAVA_OPTIONS", config[0].Name)
	elements := strings.Split(config[0].Value, "-D")
	assert.True(t, strings.HasPrefix(config[0].Value, existingOpts))
	assert.True(t, strings.Contains(elements[0], getJavaagentPath()))
	assert.Equal(t, "otel.traces.exporter=otlp ", elements[1])                                // the trailing space is expected
	assert.Equal(t, "otel.exporter.otlp.traces.endpoint=http://otlp-host:4317 ", elements[2]) // the trailing space is expected
	assert.Equal(t, "otel.service.name="+serviceName+" ", elements[3])                        // the trailing space is expected
}

func TestShouldChooseOtlpConfiguration(t *testing.T) {
	// given
	serviceName := "super-app"
	existingOpts := "some-opts"
	exporter := "otlp"
	collectorHost := "collector-host"

	// when
	config := getConfiguration(exporter, serviceName, existingOpts, collectorHost)

	// then
	assert.Equal(t, 1, len(config))
	assert.True(t, strings.Contains(config[0].Value, exporter))
}

func TestShouldFallbackToOtlpConfigurationForUnknownExporter(t *testing.T) {
	// given
	serviceName := "super-app"
	existingOpts := "some-opts"
	exporter := "unknown"
	collectorHost := "collector-host"

	// when
	config := getConfiguration(exporter, serviceName, existingOpts, collectorHost)

	// then
	assert.Equal(t, 1, len(config))
	assert.True(t, strings.Contains(config[0].Value, "otlp"))
}
