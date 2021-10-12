#!/bin/bash
set -e

VERSION="1.6.2"

AWS_REGION="us-east-1"
# TODO Change "a4t4y2n3" to "sumologic" after alias migration in the AWS ECR
REGISTRY="public.ecr.aws/a4t4y2n3"
REPOSITORY="opentelemetry-java-instrumentation-jar"

VERSION_TAG="${REGISTRY}/${REPOSITORY}:${VERSION}"
LATEST_TAG="${REGISTRY}/${REPOSITORY}:latest"

echo "Login to AWS ECR"
aws ecr-public get-login-password --region ${AWS_REGION} \
	| docker login --username AWS --password-stdin ${REGISTRY}

echo "Building version ${VERSION}"
docker build --build-arg INSTRUMENTATION_VERSION=${VERSION} -t "${VERSION_TAG}" .

echo "Pushing version ${VERSION} to docker hub (assuming that you are logged in)"
docker push "${VERSION_TAG}"

echo "Pushing version 'latest' to docker hub"
docker tag "${VERSION_TAG}" "${REPOSITORY}:latest"
docker push "${LATEST_TAG}"

echo "Done. Please remember to update both version in this script and in README"
