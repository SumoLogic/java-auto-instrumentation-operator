# Introduction
If you read this you probably want to publish a new version of OT jars image.
In this document you can find instructions how to publish the image.

# Location
The image is hosted in Docker Hub in sumologic organization https://hub.docker.com/r/sumologic/opentelemetry-jars 

# Version history
- `v0.1.0` - contains OT jars version `0.3.0`
- `v0.2.0` - contains OT jars version `0.6.0`

# Before you start building
- Please download OT jars you want to include in the image. Make sure that there
are no other jar files in this directory.
- Make sure that `VERSION` in `image-build-and-push.sh` refers to the version 
    you want to release.
- Make sure you are logged into Docker Hub and have write permissions to sumologic organization

# Build & publish
Just run the `image-build-and-push.sh` script.
