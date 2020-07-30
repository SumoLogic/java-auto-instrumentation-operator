# java-auto-instrumentation-operator

This repo contains Kubernetes operator that allows auto instrumentation of Java applications.

# Installation
The best way to install the operator is to use helm chart:
```shell script
helm repo add java-auto-instrumentation-operator https://sumologic.github.io/java-auto-instrumentation-operator
helm install java-auto-instrumentation-operator/java-auto-instrumentation-operator --name operator --namespace operator-helm
```
This is helm 3 chart.
In the example presented above I've assumed you want to install the operator in the `operator-helm` namespace
of your Kubernetes cluster.

# License
TBD

# Contributing
TBD

