# java-auto-instrumentation-operator

This repo contains Kubernetes operator that allows auto instrumentation of Java applications. It does so by automatically injecting Java Auto Instrumentation agent and configuration

# Installation

## Preparing the pods

The deployment needs to have following labels applied:
* `should-auto-instrument` set to *true*
* `auto-instrumentation-exporter` preferably set to `otlp`
* `auto-instr-service-name` set to name the service should be presented in spans
* `collector-host` set to host where spans need to be sent

For example:

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: service-a
  namespace: java-demo-app
  labels:
    should-auto-instrument: "true"
    auto-instrumentation-exporter: "otlp"
    auto-instr-service-name: "service-a"
    collector-host: "collection-sumologic-otelcol.sumologic"
```

## Adding the Operator

The best way to install the operator is to install it as a Helm chart. For example, to install it in `opeartor-helm` namespace, run
following commands (for Helm 3):

```shell script
helm repo add java-auto-instrumentation-operator https://sumologic.github.io/java-auto-instrumentation-operator
helm install operator java-auto-instrumentation-operator/java-auto-instrumentation-operator --namespace operator-helm
```

# Limitations

Currently only single-container deployments are being supported

Following operator installation, auto-instrumentation injection is applied only for newly started or restarted pods (currently running services need to be restarted manually to enable auto-instrumentation for them).

# License
Apache 2

