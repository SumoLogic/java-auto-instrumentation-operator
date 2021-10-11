# Introduction
If you read this you probably want to publish a new version of [OpenTelemetry Java Instrumentation](https://github.com/open-telemetry/opentelemetry-java-instrumentation) jar image.
In this document you can find instructions how to publish the image.

# Location
The image is hosted in AWS ECR in a4t4y2n3 (sumologic) organization https://gallery.ecr.aws/a4t4y2n3/opentelemetry-java-instrumentation-jar.

# Version history
- `1.6.2`  - contains OpenTelemetry Java Instrumentation JAR version `1.6.2`

# Before you start building
- Make sure that `VERSION` in `image-build-and-push.sh` refers to the version 
    you want to release.
- Make sure you are logged into Sumo Logic ECR and have `push` permissions.

# Build & publish
Just run the `image-build-and-push.sh` script.
