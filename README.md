# java-auto-instrumentation-operator

This repo contains Kubernetes operator that allows auto instrumentation of Java applications.

# Installation

## Preparing the pods

The deployment needs to have following labels applied:
* `should-auto-instrument` set to *true*
* `auto-instrumentation-exporter` preferably set to `otlp`
* `auto-instr-service-name` set to name the service should be presented in spans

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
```

## Adding the Operator

The best way to install the operator is to use helm chart:
```shell script
helm repo add java-auto-instrumentation-operator https://sumologic.github.io/java-auto-instrumentation-operator
helm install java-auto-instrumentation-operator/java-auto-instrumentation-operator --name operator --namespace operator-helm
```
This is helm 3 chart.
In the example presented above I've assumed you want to install the operator in the `operator-helm` namespace
of your Kubernetes cluster.

# Limitations

Currently only single-container deployments are being supported

# License
Apache 2

# Contributing
TBD

