#!/usr/local/bin/bash
set -e

VERSION="1.6.2"
AWS_ACCOUNT_NUMBER="663229565520"
AWS_REGION=us-west-2
REGISTRY="${AWS_ACCOUNT_NUMBER}.dkr.ecr.${AWS_REGION}.amazonaws.com"

# TODO Change "a4t4y2n3" to "sumologic" after alias migration in the AWS ECR
REPOSITORY="a4t4y2n3/opentelemetry-java-instrumentation-jar"

VERSION_TAG="${REGISTRY}/${REPOSITORY}:${VERSION}"
LATEST_TAG="${REGISTRY}/${REPOSITORY}:latest"

echo "Login to AWS ECR"
aws ecr get-login-password --region ${AWS_REGION} | docker login --username AWS --password-stdin ${REGISTRY}

echo "Building version ${VERSION}"
docker build --build-arg INSTRUMENTATION_VERSION=${VERSION} -t "${VERSION_TAG}" .

echo "Pushing version ${VERSION} to docker hub (assuming that you are logged in)"
docker push "${VERSION_TAG}"

echo "Pushing version 'latest' to docker hub"
docker tag "${VERSION_TAG}" "${REPOSITORY}:latest"
docker push "${LATEST_TAG}"

echo "Done. Please remember to update both version in this script and in README"
