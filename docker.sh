#!/bin/bash
# Build the docker image
# -t Tags the image with the name dockerize
docker image build -f Dockerfile -t dockerize .
# -f specifies the path
# Run the docker container from the image made above
docker container run -p 8080:8080 --name forum dockerize