package javaautoinstrumentationoperator

import (
	"strings"

	"github.com/stretchr/testify/assert"

	"testing"
)

func TestShouldUseJaegerExporterExplicitly(t *testing.T) {
	// given
	deployment := buildDeployment(map[string]string{
		"sumo-traces-exporter": "jaeger",
	})

	// when
	exporter := getTracesExporterOrDefault(testLogger, deployment)

	// then
	assert.Equal(t, "jaeger", exporter)
}

func TestShouldBuildJaegerConfiguration(t *testing.T) {
	// given
	serviceName := "super-app"
	existingOpts := "some-opts"
	collectorHost := "jaeger-host"

	// when
	config := getJaegerConfiguration(existingOpts, collectorHost, serviceName)

	// then
	assert.Equal(t, 1, len(config))

	assert.Equal(t, "_JAVA_OPTIONS", config[0].Name)
	elements := strings.Split(config[0].Value, "-D")
	assert.True(t, strings.HasPrefix(config[0].Value, existingOpts))
	assert.True(t, strings.Contains(elements[0], getJavaagentPath()))
	assert.Equal(t, "otel.traces.exporter=jaeger ", elements[1])                            // the trailing space is expected
	assert.Equal(t, "otel.exporter.jaeger.endpoint=http://jaeger-host:14250 ", elements[2]) // the trailing space is expected
	assert.Equal(t, "otel.service.name=super-app ", elements[3])                            // the trailing space is expected
}

func TestShouldChooseJaegerConfiguration(t *testing.T) {
	// given
	serviceName := "super-app"
	existingOpts := "some-opts"
	exporter := "jaeger"

	// when
	config := getConfiguration(exporter, serviceName, existingOpts, exporter)

	// then
	assert.Equal(t, 1, len(config))
	assert.True(t, strings.Contains(config[0].Value, exporter))
}
