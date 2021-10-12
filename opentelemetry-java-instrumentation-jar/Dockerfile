FROM busybox:latest

WORKDIR /ot-jar

ARG INSTRUMENTATION_VERSION
ENV INSTRUMENTATION_VERSION ${INSTRUMENTATION_VERSION:-1.6.2}

ENV INSTRUMENTATION_JAR="https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/download/\
v${INSTRUMENTATION_VERSION}/opentelemetry-javaagent-all.jar"

RUN wget ${INSTRUMENTATION_JAR}
