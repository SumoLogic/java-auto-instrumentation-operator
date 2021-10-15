package javaautoinstrumentationoperator

import (
	"fmt"

	"github.com/go-logr/logr"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func getTracesExporterOrDefault(reqLogger logr.Logger, deployment *appv1.Deployment) string {
	exporter, ok := deployment.Labels[opentelemetryTracesExporterLabel]
	if ok {
		if exporter == "jaeger" || exporter == "otlp" || exporter == "zipkin" {
			return exporter
		} else {
			reqLogger.Info("Unknown exporter "+exporter+", will default to OTLP/gRPC", "Deployment",
				deployment.Name)
			return "otlp"
		}
	} else {
		reqLogger.Info("No exporter set, will default to OTLP/gRPC", "Deployment",
			deployment.Name)
		return "otlp"
	}
}

func getJavaagentPath() string {
	return fmt.Sprintf(" -javaagent:%s ", opentelemetryJarFinalPath)
}

func getExporterOpt(exporterName string) string {
	const exporterOpt = "-Dotel.traces.exporter"

	return fmt.Sprintf("%s=%s", exporterOpt, exporterName)
}

func getServiceNameOpt(serviceName string) string {
	const serviceNameOpt = "-Dotel.service.name"

	return fmt.Sprintf("%s=%s", serviceNameOpt, serviceName)
}

func getJaegerConfiguration(existingJavaOptions string, collectorHost string, serviceName string) []corev1.EnvVar {
	const exporter = "jaeger"
	exporterOpt := getExporterOpt(exporter)
	exporterOptEndpoint := fmt.Sprintf("-Dotel.exporter.jaeger.endpoint=http://%s:14250", collectorHost)
	serviceNameOpt := getServiceNameOpt(serviceName)
	javaAgentPath := getJavaagentPath()

	envValue := fmt.Sprintf("%s%s %s %s %s ", existingJavaOptions, javaAgentPath, exporterOpt, exporterOptEndpoint, serviceNameOpt)

	return []corev1.EnvVar{
		{
			Name:  javaOptionsEnvVar,
			Value: envValue,
		},
	}
}

func getZipkinConfiguration(existingJavaOptions string, collectorHost string, serviceName string) []corev1.EnvVar {
	const exporter = "zipkin"
	exporterOpt := getExporterOpt(exporter)
	exporterOptEndpoint := fmt.Sprintf("-Dotel.exporter.zipkin.endpoint=http://%s:9411/api/v2/spans", collectorHost)
	serviceNameOpt := getServiceNameOpt(serviceName)
	javaAgentPath := getJavaagentPath()

	envValue := fmt.Sprintf("%s%s %s %s %s ", existingJavaOptions, javaAgentPath, exporterOpt, exporterOptEndpoint, serviceNameOpt)

	return []corev1.EnvVar{
		{
			Name:  javaOptionsEnvVar,
			Value: envValue,
		},
	}
}

func getTracesOtlpConfiguration(existingJavaOptions string, collectorHost string, serviceName string) []corev1.EnvVar {
	const exporter = "otlp"
	exporterOpt := getExporterOpt(exporter)
	exporterOptEndpoint := fmt.Sprintf("-Dotel.exporter.otlp.traces.endpoint=http://%s:4317", collectorHost)
	serviceNameOpt := getServiceNameOpt(serviceName)
	javaAgentPath := getJavaagentPath()

	envValue := fmt.Sprintf("%s%s %s %s %s ", existingJavaOptions, javaAgentPath, exporterOpt, exporterOptEndpoint, serviceNameOpt)

	return []corev1.EnvVar{
		{
			Name:  javaOptionsEnvVar,
			Value: envValue,
		},
	}
}
