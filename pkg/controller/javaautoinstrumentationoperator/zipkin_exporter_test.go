package javaautoinstrumentationoperator

import (
	"strings"

	"github.com/stretchr/testify/assert"

	"testing"
)

func TestShouldUseZipkinExporterExplicitly(t *testing.T) {
	// given
	deployment := buildDeployment(map[string]string{
		"sumo-traces-exporter": "zipkin",
	})

	// when
	exporter := getTracesExporterOrDefault(testLogger, deployment)

	// then
	assert.Equal(t, "zipkin", exporter)
}

func TestShouldBuildZipkinConfiguration(t *testing.T) {
	// given
	serviceName := "super-app"
	existingOpts := "some-opts"
	collectorHost := "zipkin-host"

	// when
	config := getZipkinConfiguration(existingOpts, collectorHost, serviceName)

	// then
	assert.Equal(t, 1, len(config))

	assert.Equal(t, "_JAVA_OPTIONS", config[0].Name)
	elements := strings.Split(config[0].Value, "-D")
	assert.True(t, strings.HasPrefix(config[0].Value, existingOpts))
	assert.True(t, strings.Contains(elements[0], getJavaagentPath()))
	assert.Equal(t, "otel.traces.exporter=zipkin ", elements[1])                                        // the trailing space is expected
	assert.Equal(t, "otel.exporter.zipkin.endpoint=http://zipkin-host:9411/api/v2/spans ", elements[2]) // the trailing space is expected
	assert.Equal(t, "otel.service.name=super-app ", elements[3])                                        // the trailing space is expected
}

func TestShouldChooseZipkinConfiguration(t *testing.T) {
	// given
	serviceName := "super-app"
	existingOpts := "some-opts"
	exporter := "zipkin"

	// when
	config := getConfiguration(exporter, serviceName, existingOpts, exporter)

	// then
	assert.Equal(t, 1, len(config))
	assert.True(t, strings.Contains(config[0].Value, exporter))
}
