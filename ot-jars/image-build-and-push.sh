#!/usr/local/bin/bash
set -e

VERSION="v0.2.0"

echo "Building version $VERSION"
docker build -t "sumologic/opentelemetry-jars:$VERSION" .

docker tag "sumologic/opentelemetry-jars:$VERSION" "sumologic/opentelemetry-jars:latest"

echo "Pushing version $VERSION to docker hub (assuming that you are logged in)"
docker push "sumologic/opentelemetry-jars:$VERSION"

echo "Pushing version 'latest' to docker hub"
docker push "sumologic/opentelemetry-jars:latest"

echo "Done. Please remember to update both version in this script and in README"
