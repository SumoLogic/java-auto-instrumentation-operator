# Introduction
This directory contains some examples and auxiliary resources that can be
helpful in deploying and debugging the operator.

# Traces
In order to see traces from Java application you must first install some services in
your cluster. There are two ways to see traces:
- in OpenTelemetry collector logs,
- in Jaeger UI.

## How to install the operator
The best way to install the operator is to use helm chart. Please take a look
at the main README of this project.

## UI for traces
This step is not required but if you want to see traces in UI you have to deploy Jaeger:
```shell script
kubectl -n your-namespace create -f jaeger.yaml
``` 
And then port forward Jaeger UI port
````shell script
kubectl -n your-namespace port-forward jaeger-pod 16686:16686
````
And then open `http://localhost:16686` in your web browser.

## Collector
Now we need something that collects and processes traces. We use OpenTelemetry
collector for that. It can be installed in the following way:
```shell script
kubectl -n your-namespace create -f otel-col.yaml
```
In order to see your traces in logs you should run the following script:
```shell script
kubectl -n your-namespace logs otel-col-pod
```

## Java application
Java application should be installed as the last one.
```shell script
kubectl -n your-namespace create -f java-app-otlp.yaml
```

## Generating traffic
In order to see traces you have to make the Java application do some work.
The best way to do that is to forward the port and then make a request:
```shell script
kubectl -n your-namespace port-forward java-app-pod 8080:8080
# in a separate shell
curl localhost:8080/info
```
You should see a JSON response with some metadata.
Now you should be able to see traces using one of the method described above.

# Miscellaneous
## How to update operator image
I'm assuming here that the operator was deployed by `operator.yaml`:
```shell script
kubectl -n your-namespace set image deployments/java-auto-instrumetation-operator java-auto-instrumentation-operator=sumologic/opentelemetry-collector-operator:v0.2.0
```
The `operator.yaml` file is also used by the helm chart.
