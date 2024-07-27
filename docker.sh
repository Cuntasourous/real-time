#!/bin/bash

# Build the docker image
# -t Tags the image with the name dockerize
docker image build -f Dockerfile -t forum-app .

# -f specifies the path
# docker image build -f Dockerfile -t dockerize .

# Run the docker container from the image made above
docker container run -p 8080:8080 --name forum-app